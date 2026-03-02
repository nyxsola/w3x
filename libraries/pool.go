package libraries

import (
	"math/big"

	"github.com/aicora/go-uniswap/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	// ErrPoolAlreadyInitialized is returned if Initialize is called twice.
	ErrPoolAlreadyInitialized = errors.New("pool already initialized")

	// ErrPoolNotInitialized is returned if state is accessed before initialization.
	ErrPoolNotInitialized = errors.New("pool not initialized")

	// ErrNoLiquidity is returned when liquidity-dependent operations are executed
	// while active liquidity is zero.
	ErrNoLiquidity = errors.New("no liquidity")

	// ErrTickLiquidityOverflow is returned when liquidity would exceed uint128 max (2^128-1).
	ErrTickLiquidityOverflow = errors.New("tick liquidity overflow")

	// ErrSwapFeeTooHigh is returned when the swap fee is too high.
	ErrSwapFeeTooHigh = errors.New("swap fee too high")

	ErrSqrtPriceLimitExceeded = errors.New("sqrt price limit exceeded")
)

// Slot0 contains the hot-path pool state.
//
// In Solidity this struct is tightly packed into a single storage slot.
// Here it models the same economic state in Go.
//
// Invariants:
//
//   - SqrtPriceX96 and Tick must be consistent via TickMath
//   - SqrtPriceX96 > 0 once initialized
//   - Fee configuration must respect protocol constraints
//
// SqrtPriceX96 uses Q64.96 fixed-point format.
type Slot0 struct {
    SqrtPriceX96 *big.Int
    Tick         int
    ProtocolFee  utils.ProtocolFee
    LPFee        utils.LPFee
}

// Pool implements the concentrated liquidity AMM state machine.
//
// It maintains:
//
//   1. Global fee accumulators (monotonic increasing)
//   2. Active in-range liquidity
//   3. Tick boundary fee accounting
//
// Financial invariants:
//
//   - feeGrowthGlobal{0,1}X128 are strictly non-decreasing
//   - liquidity >= 0
//   - liquidityGross(T) >= 0 for any tick T
//   - Crossing a tick applies liquidityNet exactly once
//
// Fee accounting follows the model introduced in Uniswap v4.
type Pool struct {
	slot0 Slot0

	// Monotonic global fee growth accumulators (Q128 precision).
	feeGrowthGlobal0X128 *big.Int
	feeGrowthGlobal1X128 *big.Int

	// Active liquidity within current price range.
	liquidity *big.Int

	// Externalized tick storage.
	tickManager ITickManager

	// Externalized position storage.
	positionManager *PositionManager
}

// NewPool creates an uninitialized pool instance.
func NewPool() *Pool {
	return &Pool {
		feeGrowthGlobal0X128: big.NewInt(0),
		feeGrowthGlobal1X128: big.NewInt(0),
		liquidity: big.NewInt(0),
		tickManager: NewTickManager(),
		positionManager: NewPositionManager(),
	}
}

// Initialize sets the initial price and LP fee.
//
// It computes Tick from sqrtPriceX96 and sets Slot0.
// Can only be executed once.
func (p *Pool) Initialize(sqrtPriceX96 *big.Int, lpFee utils.LPFee) (int, error) {
	if p.slot0.SqrtPriceX96 != nil && p.slot0.SqrtPriceX96.Cmp(big.NewInt(0)) != 0 {
		return 0, ErrPoolAlreadyInitialized
	}

	tick, err := utils.GetTickAtSqrtPrice(sqrtPriceX96)
	if err != nil {
		return 0, err
	}

	p.slot0 = Slot0{
		SqrtPriceX96: new(big.Int).Set(sqrtPriceX96),
		Tick:         tick,
		LPFee:        lpFee,
		ProtocolFee:  0,
	}

	return tick, nil
}

// CheckPoolInitialized verifies pool has been initialized.
func (p *Pool) CheckPoolInitialized() error {
	if p.slot0.SqrtPriceX96 == nil || p.slot0.SqrtPriceX96.Cmp(big.NewInt(0)) == 0 {
		return ErrPoolNotInitialized
	}
	return nil
}

// SetProtocolFee updates protocol fee configuration.
func (p *Pool) SetProtocolFee(protocolFee utils.ProtocolFee) error {
	if err := p.CheckPoolInitialized(); err != nil {
		return err
	}

	p.slot0.ProtocolFee = protocolFee
	return nil
}

// SetLPFee updates liquidity provider fee.
func (p *Pool) SetLPFee(lpFee utils.LPFee) error {
	if err := p.CheckPoolInitialized(); err != nil {
		return err
	}

	p.slot0.LPFee = lpFee
	return nil
}

// ClearTick removes tick storage when liquidityGross becomes zero.
func (p *Pool) ClearTick(tick int) {
	p.tickManager.Clear(tick)
}

// GetFeeGrowthInside computes cumulative fee growth inside [tickLower, tickUpper).
//
// Tick axis:
// 
// 0    tickLower        tickUpper     MaxTick
// |-------|-----------------|-------------|
//    10           20              70
//
// | tickCurrent position                 | formula                                         | conceptual result            
// | ------------------------------------ | ----------------------------------------------- | ------------------------- 
// | tickCurrent < tickLower              | lower.FeeGrowthOutside - upper.FeeGrowthOutside | 10 - 70 → negative (not started)   
// | tickLower <= tickCurrent < tickUpper | feeGrowthGlobal - lower - upper                 | 100 - 10 - 70 = 20 (internal accrual) 
// | tickCurrent >= tickUpper             | upper - lower                                   | 70 - 10 = 60 (fully accrued)   
func (p *Pool) GetFeeGrowthInside(tickLower, tickUpper int) (feeGrowthInside0X128, feeGrowthInside1X128 *big.Int) {
	lower := p.tickManager.Get(tickLower)
	upper := p.tickManager.Get(tickUpper)
	tickCurrent := p.slot0.Tick

	switch {
	case tickCurrent < tickLower:
		feeGrowthInside0X128 = new(big.Int).Sub(lower.FeeGrowthOutside0X128, upper.FeeGrowthOutside0X128)
		feeGrowthInside1X128 = new(big.Int).Sub(lower.FeeGrowthOutside1X128, upper.FeeGrowthOutside1X128)

	case tickCurrent >= tickUpper:
		feeGrowthInside0X128 = new(big.Int).Sub(upper.FeeGrowthOutside0X128, lower.FeeGrowthOutside0X128)
		feeGrowthInside1X128 = new(big.Int).Sub(upper.FeeGrowthOutside1X128, lower.FeeGrowthOutside1X128)

	default: // tickLower <= tickCurrent < tickUpper
		feeGrowthInside0X128 = new(big.Int).Sub(p.feeGrowthGlobal0X128, lower.FeeGrowthOutside0X128)
		feeGrowthInside0X128.Sub(feeGrowthInside0X128, upper.FeeGrowthOutside0X128)

		feeGrowthInside1X128 = new(big.Int).Sub(p.feeGrowthGlobal1X128, lower.FeeGrowthOutside1X128)
		feeGrowthInside1X128.Sub(feeGrowthInside1X128, upper.FeeGrowthOutside1X128)
	}

	return
}

// CrossTick executes boundary crossing logic.
//
// Transformation:
//
//   Fo(T) = G - Fo(T)
//
// This flips fee accounting perspective when
// price crosses the boundary.
//
// Returns liquidityNet to apply to active liquidity.
func (p *Pool) CrossTick(tick int, feeGrowthGlobal0X128, feeGrowthGlobal1X128 *big.Int) *big.Int {
    currentTick := p.tickManager.Get(tick)

    // feeGrowthOutside := feeGrowthGlobal - feeGrowthOutside
    currentTick.FeeGrowthOutside0X128 = new(big.Int).Sub(feeGrowthGlobal0X128, currentTick.FeeGrowthOutside0X128)
    currentTick.FeeGrowthOutside1X128 = new(big.Int).Sub(feeGrowthGlobal1X128, currentTick.FeeGrowthOutside1X128)

    return currentTick.LiquidityNet
}

// UpdateTick mutates liquidity state at a boundary.
//
// Steps:
//
// 1. Update liquidityGross
// 2. Detect flip (zero <-> non-zero)
// 3. Initialize feeGrowthOutside if first activation
// 4. Update liquidityNet
//
// liquidityNet semantics:
//   When crossing upward:
//       liquidity += liquidityNet
func (p *Pool) UpdateTick(tick int, liquidityDelta *big.Int, upper bool) (flipped bool, liquidityGrossAfter *big.Int) {
	tickInfo := p.tickManager.Get(tick)

	liquidityGrossAfter = new(big.Int).Add(tickInfo.LiquidityGross, liquidityDelta)
	flipped = (liquidityGrossAfter.Sign() == 0) != (tickInfo.LiquidityGross.Sign() == 0)

	if tickInfo.LiquidityGross.Sign() == 0 {
		if tick <= p.slot0.Tick {
			tickInfo.FeeGrowthOutside0X128.Set(p.feeGrowthGlobal0X128)
			tickInfo.FeeGrowthOutside1X128.Set(p.feeGrowthGlobal1X128)
		}
	}

	if upper {
		tickInfo.LiquidityNet.Sub(tickInfo.LiquidityNet, liquidityDelta)
	} else {
		tickInfo.LiquidityNet.Add(tickInfo.LiquidityNet, liquidityDelta)
	}

	tickInfo.LiquidityGross.Set(liquidityGrossAfter)

	return
}

// Donate distributes tokens proportionally to all active liquidity.
//
// Formula:
//
//   feeGrowthGlobal += (amount * Q128) / liquidity
//
// Properties:
//
//   - Monotonic fee accumulator
//   - No price movement
//   - No tick mutation
func (p *Pool) Donate(amount0, amount1 *big.Int) (BalanceDelta, error) {
	if p.liquidity.Sign() == 0 {
		return ZeroBalanceDelta, ErrNoLiquidity
	}

	delta := BalanceDelta{
		Amount0: new(big.Int).Neg(amount0),
		Amount1: new(big.Int).Neg(amount1),
	}

	if amount0.Sign() > 0 {
		increment := new(big.Int).Mul(amount0, utils.Q128)
		increment.Div(increment, p.liquidity)
		p.feeGrowthGlobal0X128.Add(p.feeGrowthGlobal0X128, increment)
	}

	if amount1.Sign() > 0 {
		increment := new(big.Int).Mul(amount1, utils.Q128)
		increment.Div(increment, p.liquidity)
		p.feeGrowthGlobal1X128.Add(p.feeGrowthGlobal1X128, increment)
	}

	return delta, nil
}

// ModifyLiquidityParams defines position mutation input.
type ModifyLiquidityParams struct {
	Owner         common.Address
	TickLower     int  
	TickUpper     int   
	LiquidityDelta *big.Int 
	TickSpacing   int   
	Salt          [32]byte
}

// ModifyLiquidityState captures tick mutation results.
type ModifyLiquidityState struct {
	FlippedLower           bool   
	LiquidityGrossAfterLower *big.Int 
	FlippedUpper           bool   
	LiquidityGrossAfterUpper *big.Int
}

// ModifyLiquidity updates a position's liquidity in the pool.
//
// This function mirrors the core behavior of Uniswap V4 `modifyPosition` logic.
// It handles:
//
//   1. Tick state mutation (liquidity gross / net update)
//   2. Tick bitmap flipping
//   3. Position fee growth accounting
//   4. Token amount delta calculation
//   5. Pool active liquidity update (if position is in-range)
//
// -----------------------------------------------------------------------------
// liquidityDelta semantics:
//
//   > 0  → add liquidity (user deposits tokens)
//   < 0  → remove liquidity (user withdraws tokens)
//   = 0  → collect fees only
//
// -----------------------------------------------------------------------------
// Returned values:
//
//   delta     → principal token0/token1 change from liquidity modification
//   feeDelta  → fees owed to the position
func (p *Pool) ModifyLiquidity(params ModifyLiquidityParams) (delta, feeDelta BalanceDelta, err error) {
	liquidityDelta := params.LiquidityDelta
	tickLower := params.TickLower
	tickUpper := params.TickUpper

	if err = utils.CheckTicks(tickLower, tickUpper); err != nil {
		return
	}

	var state ModifyLiquidityState

	if liquidityDelta.Sign() != 0 {
		state.FlippedLower, state.LiquidityGrossAfterLower = p.UpdateTick(tickLower, liquidityDelta, false)
		state.FlippedUpper, state.LiquidityGrossAfterUpper = p.UpdateTick(tickUpper, liquidityDelta, true)

		if liquidityDelta.Sign() > 0 {
			maxLiquidityPerTick := new(big.Int)
			maxLiquidityPerTick, err = utils.TickSpacingToMaxLiquidityPerTick(params.TickSpacing)
			if err != nil {
				return
			}
			if state.LiquidityGrossAfterLower.Cmp(maxLiquidityPerTick) > 0 {
				return BalanceDelta{}, BalanceDelta{}, errors.Wrapf(ErrTickLiquidityOverflow, "lower tick %d", tickLower)
			}
			if state.LiquidityGrossAfterUpper.Cmp(maxLiquidityPerTick) > 0 {
				return BalanceDelta{}, BalanceDelta{}, errors.Wrapf(ErrTickLiquidityOverflow, "upper tick %d", tickUpper)
			}
		}

		if state.FlippedLower {
			p.tickManager.FlipTick(tickLower, params.TickSpacing)
		}
		if state.FlippedUpper {
			p.tickManager.FlipTick(tickUpper, params.TickSpacing)
		}
	}

	feeGrowthInside0X128, feeGrowthInside1X128 := p.GetFeeGrowthInside(tickLower, tickUpper)

	position := p.positionManager.Get(params.Owner, tickLower, tickUpper, params.Salt)

	feesOwed0, feesOwed1, err := position.Update(liquidityDelta, feeGrowthInside0X128, feeGrowthInside1X128)
	if err != nil {
		return
	}
	
	feeDelta = NewBalanceDelta(feesOwed0, feesOwed1)

	if liquidityDelta.Sign() < 0 {
		if state.FlippedLower {
			p.ClearTick(tickLower)
		}
		if state.FlippedUpper {
			p.ClearTick(tickUpper)
		}
	}

	if liquidityDelta.Sign() != 0 {
		slot0 := p.slot0
		tick := slot0.Tick
		sqrtPriceX96 := slot0.SqrtPriceX96
		tickLowerSqrtPrice := big.NewInt(0)
		tickUpperSqrtPrice := big.NewInt(0)
		amount0 := big.NewInt(0)
		amount1 := big.NewInt(0)

		if tick < tickLower {
			tickLowerSqrtPrice, err = utils.GetSqrtPriceAtTick(tickLower)
			if err != nil {
				return
			}
			tickUpperSqrtPrice, err = utils.GetSqrtPriceAtTick(tickUpper)
			if err != nil {
				return
			}
			
			if liquidityDelta.Sign() < 0 {
				amount0, err = utils.GetAmount0Delta(tickLowerSqrtPrice, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), false)
				if err != nil {
					return
				}
			} else {
				amount0, err = utils.GetAmount0Delta(tickLowerSqrtPrice, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), true)
				if err != nil {
					return
				}
				amount0 = amount0.Neg(amount0)
			}
			
			delta = NewBalanceDelta(amount0, big.NewInt(0))
		} else if tick < tickUpper {
			tickUpperSqrtPrice, err = utils.GetSqrtPriceAtTick(tickUpper)
			if err != nil {
				return
			}

			if liquidityDelta.Sign() < 0 {
				amount0, err = utils.GetAmount0Delta(sqrtPriceX96, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), false)
				if err != nil {
					return
				}
			} else {
				amount0, err = utils.GetAmount0Delta(sqrtPriceX96, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), true)
				if err != nil {
					return
				}
				amount0 = amount0.Neg(amount0)
			}
			
			tickLowerSqrtPrice, err = utils.GetSqrtPriceAtTick(tickLower)
			if err != nil {
				return
			}

			if liquidityDelta.Sign() < 0 {
				amount1, err = utils.GetAmount1Delta(tickLowerSqrtPrice, sqrtPriceX96, new(big.Int).Abs(liquidityDelta), false)
				if err != nil {
					return
				}
			} else {
				amount1, err = utils.GetAmount1Delta(tickLowerSqrtPrice, sqrtPriceX96, new(big.Int).Abs(liquidityDelta), true)
				if err != nil {
					return
				}
				amount1 = amount1.Neg(amount1)
			}

			delta = NewBalanceDelta(amount0, amount1)

			p.liquidity.Add(p.liquidity, liquidityDelta)
		} else {
			tickLowerSqrtPrice, err = utils.GetSqrtPriceAtTick(tickLower)
			if err != nil {
				return
			}
			tickUpperSqrtPrice, err = utils.GetSqrtPriceAtTick(tickUpper)
			if err != nil {
				return
			}

			if liquidityDelta.Sign() < 0 {
				amount1, err = utils.GetAmount1Delta(tickLowerSqrtPrice, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), false)
				if err != nil {
					return
				}
			} else {
				amount1, err = utils.GetAmount1Delta(tickLowerSqrtPrice, tickUpperSqrtPrice, new(big.Int).Abs(liquidityDelta), true)
				if err != nil {
					return
				}
				amount1 = amount1.Neg(amount1)
			}

			delta = NewBalanceDelta(big.NewInt(0), amount1)
		}
	}

	return
}

// SwapResult mirrors the final state returned after a swap completes.
//
// Equivalent to Solidity SwapResult.
type SwapResult struct {
	// Current sqrt price (Q64.96)
	SqrtPriceX96 *big.Int

	// Current tick corresponding to sqrt price
	Tick int

	// Active in-range liquidity
	Liquidity *big.Int
}

// StepComputations represents per-step swap calculations
// while iterating across ticks.
type StepComputations struct {

	// sqrt price at the beginning of this step
	SqrtPriceStartX96 *big.Int

	// Next tick in swap direction
	TickNext int

	// Whether the next tick is initialized
	Initialized bool

	// sqrt price at next tick
	SqrtPriceNextX96 *big.Int

	// Amount of input token consumed in this step
	AmountIn *big.Int

	// Amount of output token produced in this step
	AmountOut *big.Int

	// Fee paid in this step
	FeeAmount *big.Int

	// Global fee growth accumulator for input token
	// Updated in storage at end of swap
	FeeGrowthGlobalX128 *big.Int
}

// SwapParams defines swap input configuration.
type SwapParams struct {

	// Amount specified:
	AmountSpecified *big.Int

	// Tick spacing for the pool
	TickSpacing int

	// Swap direction:
	// true  -> token0 → token1
	// false -> token1 → token0
	ZeroForOne bool

	// Price limit boundary (Q64.96)
	SqrtPriceLimitX96 *big.Int

	// Optional LP fee override (in hundredths of a bip, e.g. 3000 = 0.3%)
	LpFeeOverride utils.LPFee
}

// Swap executes a token swap within the Pool according to the provided parameters.
//
// The function performs a multi-step swap calculation, updating the pool's state
// (tick, sqrtPriceX96, liquidity, and fee growth) according to the swap direction
// and specified limits. It supports both exact input and exact output swaps,
// handles protocol fees, and accumulates fees for liquidity providers.
//
// Parameters:
//   - params: SwapParams containing the following:
//       * ZeroForOne: direction of the swap (true if swapping token0 for token1)
//       * AmountSpecified: the exact amount in or out for the swap
//       * SqrtPriceLimitX96: the boundary price for the swap (swap stops if crossed)
//       * TickSpacing: tick spacing of the pool (used for tick crossing calculations)
//       * LpFeeOverride: optional LP fee override (validated if set)
//
// Returns:
//   - swapDelta: BalanceDelta representing net token deltas for the swap
//   - amountToProtocol: the accumulated protocol fee in token units
//   - swapFee: effective swap fee rate applied for this swap
//   - result: SwapResult representing the final pool state after the swap
//   - err: any error encountered (e.g., invalid parameters, fee limits exceeded)
//
// Notes:
//   - This function does not revert the pool state on partial failures;
//     careful error handling is required at the caller side.
//   - All arithmetic is performed with *big.Int to support high-precision calculations.
//   - The function ensures that SqrtPrice limits are respected to prevent over-slippage.
//   - Protocol fee and LP fee are correctly split and accumulated according to Uniswap v3/v4 model.
//   - Tick crossing and liquidity updates are performed only when initialized ticks are crossed.
//
// Implementation Details:
//   1. Copy current pool slot0 state to local variables for calculation safety.
//   2. Determine swap direction (zeroForOne) and fetch protocol fees.
//   3. Initialize amount remaining, amount calculated, and swap result snapshot.
//   4. Compute effective swap fee (protocol + LP fee), validating against MaxSwapFee.
//   5. Perform boundary checks for SqrtPriceLimitX96 to prevent invalid swaps.
//   6. Initialize StepComputations struct for iterative swap step processing.
//   7. Loop over swap steps until either amountSpecifiedRemaining == 0 or price limit reached:
//       a. Determine next initialized tick within one word (tick crossing logic)
//       b. Compute sqrtPriceNextX96 from TickNext
//       c. Perform ComputeSwapStep to calculate amountIn, amountOut, and feeAmount
//       d. Update amountSpecifiedRemaining and amountCalculated based on swap direction
//       e. Deduct protocol fees from step.FeeAmount and accumulate in amountToProtocol
//       f. Update fee growth globals for LPs
//       g. Handle tick crossing and adjust result.Tick and result.Liquidity
//   8. After loop, update pool slot0 and liquidity state with final result.
//   9. Compute final swapDelta based on exact input/output direction.
//
// Errors:
//   - ErrSwapFeeTooHigh: if computed swap fee exceeds MaxSwapFee
//   - ErrSqrtPriceLimitExceeded: if swap attempts to cross specified sqrtPrice limits
//   - Errors from tickManager, ComputeSwapStep, or MulDiv arithmetic operations
func (p *Pool) Swap(params SwapParams) (swapDelta BalanceDelta, amountToProtocol *big.Int, swapFee uint32, result SwapResult, err error) {
	slot0Start := p.slot0
	zeroForOne := params.ZeroForOne

	amountToProtocol = new(big.Int)

	var protocolFee uint16
	if zeroForOne {
		protocolFee = slot0Start.ProtocolFee.ZeroForOne()
	} else {
		protocolFee = slot0Start.ProtocolFee.OneForZero()
	}

	amountSpecifiedRemaining := new(big.Int).Set(params.AmountSpecified)
	amountCalculated := big.NewInt(0)

	result = SwapResult{
		SqrtPriceX96: new(big.Int).Set(slot0Start.SqrtPriceX96),
		Tick:         slot0Start.Tick,
		Liquidity:    new(big.Int).Set(p.liquidity),
	}

	lpFee := slot0Start.LPFee

	if params.LpFeeOverride.IsOverride()  {
		lpFee, err = params.LpFeeOverride.RemoveOverrideFlagAndValidate()
		if err != nil {
			return
		}
	}
	
	if protocolFee == 0 {
		swapFee = uint32(lpFee)
	} else {
		swapFee = utils.CalculateSwapFee(protocolFee, lpFee)
	}

	if big.NewInt(int64(swapFee)).Cmp(utils.MaxSwapFee) > 0 {
		if params.AmountSpecified.Sign() > 0 {
			return BalanceDelta{}, nil, 0, SwapResult{}, ErrSwapFeeTooHigh
		}
	}

	if params.AmountSpecified.Sign() == 0 {
		return
	}

	if zeroForOne {
		if params.SqrtPriceLimitX96.Cmp(slot0Start.SqrtPriceX96) >= 0 {
			return BalanceDelta{}, nil, 0, SwapResult{}, ErrSqrtPriceLimitExceeded
		}
		if params.SqrtPriceLimitX96.Cmp(utils.MinSqrtPrice) <=0 {
			return BalanceDelta{}, nil, 0, SwapResult{}, ErrSqrtPriceLimitExceeded
		}
	} else {
		if params.SqrtPriceLimitX96.Cmp(slot0Start.SqrtPriceX96) <= 0 {
			return BalanceDelta{}, nil, 0, SwapResult{}, ErrSqrtPriceLimitExceeded
		}
		if params.SqrtPriceLimitX96.Cmp(utils.MaxSqrtPrice) >= 0 {
			return BalanceDelta{}, nil, 0, SwapResult{}, ErrSqrtPriceLimitExceeded
		}
	}

	step := StepComputations{
		SqrtPriceStartX96: big.NewInt(0),
		TickNext: 0,
		Initialized: false,
		SqrtPriceNextX96: big.NewInt(0),
		AmountIn: big.NewInt(0),
		AmountOut: big.NewInt(0),
		FeeAmount: big.NewInt(0),
		FeeGrowthGlobalX128: big.NewInt(0),
	}

	if zeroForOne {
		step.FeeGrowthGlobalX128 = new(big.Int).Set(p.feeGrowthGlobal0X128)
	} else {
		step.FeeGrowthGlobalX128 = new(big.Int).Set(p.feeGrowthGlobal1X128)
	}

	for amountSpecifiedRemaining.Sign() != 0 && result.SqrtPriceX96.Cmp(params.SqrtPriceLimitX96) != 0 {
		step.SqrtPriceStartX96.Set(result.SqrtPriceX96)

		step.TickNext, step.Initialized, err = p.tickManager.NextInitializedTickWithinOneWord(result.Tick, params.TickSpacing, zeroForOne)
		if err != nil {
			return
		}
		
		if step.TickNext <= utils.MinTick {
			step.TickNext = utils.MinTick
		}
		if step.TickNext >= utils.MaxTick {
			step.TickNext = utils.MaxTick
		}

		step.SqrtPriceNextX96, err = utils.GetSqrtPriceAtTick(step.TickNext);
		if err != nil {
			return
		}

		result.SqrtPriceX96, step.AmountIn, step.AmountOut, step.FeeAmount, err = utils.ComputeSwapStep(
                result.SqrtPriceX96,
                utils.GetSqrtPriceTarget(zeroForOne, step.SqrtPriceNextX96, params.SqrtPriceLimitX96),
                result.Liquidity,
                amountSpecifiedRemaining,
                swapFee,
            );
		if err != nil {
			return
		}

		if params.AmountSpecified.Sign() > 0 {
			amountSpecifiedRemaining.Sub(amountSpecifiedRemaining, step.AmountOut)
			tmp := new(big.Int).Add(step.AmountIn, step.FeeAmount)
			amountCalculated.Sub(amountCalculated, tmp)
		} else {
			tmp := new(big.Int).Add(step.AmountIn, step.FeeAmount)
			amountSpecifiedRemaining.Add(amountSpecifiedRemaining, tmp)
			amountCalculated.Add(amountCalculated, step.AmountOut)
		}

		if protocolFee > 0 {
			var delta *big.Int
			if swapFee == uint32(protocolFee) {
				delta = new(big.Int).Set(step.FeeAmount)
			} else {
				totalIn := new(big.Int).Add(step.AmountIn, step.FeeAmount)
				delta = new(big.Int).Mul(totalIn, big.NewInt(int64(protocolFee)))
				delta.Div(delta, big.NewInt(int64(utils.PipsDenominator)))
			}
			step.FeeAmount.Sub(step.FeeAmount, delta)
			amountToProtocol.Add(amountToProtocol, delta)
		}

		if result.Liquidity.Sign() > 0 {
			increment, err := utils.MulDiv(step.FeeAmount, utils.Q128, result.Liquidity)
			if err != nil {
				return BalanceDelta{}, nil, 0, SwapResult{}, err
			}
			step.FeeGrowthGlobalX128.Add(step.FeeGrowthGlobalX128, increment)
		}

		if result.SqrtPriceX96.Cmp(step.SqrtPriceNextX96) == 0 {
			if step.Initialized {
				feeGrowthGlobal0X128 := big.NewInt(0)
				feeGrowthGlobal1X128 := big.NewInt(0)

				if zeroForOne {
					feeGrowthGlobal0X128, feeGrowthGlobal1X128 = step.FeeGrowthGlobalX128, p.feeGrowthGlobal1X128
				} else {
					feeGrowthGlobal0X128, feeGrowthGlobal1X128 = p.feeGrowthGlobal0X128, step.FeeGrowthGlobalX128
				}
				
				liquidityNet := p.CrossTick(step.TickNext, feeGrowthGlobal0X128, feeGrowthGlobal1X128)
				
				if zeroForOne {
					liquidityNet = new(big.Int).Neg(liquidityNet)
				}

				result.Liquidity = utils.AddDelta(result.Liquidity, liquidityNet);
			}

			if zeroForOne {
				result.Tick = step.TickNext - 1
			} else {
				result.Tick = step.TickNext
			}

		} else if result.SqrtPriceX96.Cmp(step.SqrtPriceStartX96) != 0 {
			result.Tick, err = utils.GetTickAtSqrtPrice(result.SqrtPriceX96)
			if err != nil {
				return
			}
		}
	}

	p.slot0.Tick = result.Tick
	p.slot0.SqrtPriceX96.Set(result.SqrtPriceX96)

	if p.liquidity.Cmp(result.Liquidity) != 0 {
		p.liquidity = result.Liquidity
	}

	if zeroForOne {
		p.feeGrowthGlobal0X128 = step.FeeGrowthGlobalX128
	} else {
		p.feeGrowthGlobal1X128 = step.FeeGrowthGlobalX128
	}

	diff := new(big.Int).Sub(params.AmountSpecified, amountSpecifiedRemaining)

	if zeroForOne != (params.AmountSpecified.Sign() < 0) {
		swapDelta = NewBalanceDelta(amountCalculated, diff)
	} else {
		swapDelta = NewBalanceDelta(diff, amountCalculated)
	}

	return
}
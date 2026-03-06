package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/aicora/w3x/amm/libraries"
	"github.com/aicora/w3x/amm/types"
	"github.com/aicora/w3x/amm/utils"
)

var (
	// ErrTickSpacingTooLarge is returned when the tick spacing exceeds the allowed maximum.
	ErrTickSpacingTooLarge = errors.New("tick spacing too large")

	// ErrTickSpacingTooSmall is returned when the tick spacing is below the allowed minimum.
	ErrTickSpacingTooSmall = errors.New("tick spacing too small")

	// ErrCurrenciesOutOfOrder is returned when Currency0 is not less than Currency1.
	ErrCurrenciesOutOfOrder = errors.New("currencies out of order")

	// ErrPoolAlreadyInitialized is returned when attempting to initialize a pool that already exists.
	ErrPoolAlreadyInitialized = errors.New("pool already initialized")

	// ErrPoolNotInitialized is returned when attempting to access a pool that has not been initialized.
	ErrPoolNotInitialized = errors.New("pool not initialized")

	// ErrSwapAmountCannotBeZero is returned when the swap amount is zero.
	ErrSwapAmountCannotBeZero = errors.New("swap amount cannot be zero")
)

// PoolManager manages multiple Uniswap pools.
// It is responsible for pool creation, initialization, and indexing.
type PoolManager struct {
	// pools maps a PoolId to its corresponding Pool instance.
	pools map[libraries.PoolId]*libraries.Pool
	// a mapping of account and currency pairs to their transient balance deltas
	currencyDelta *libraries.CurrencyDelta
	// accrual of protocol fees
	protocolFeesAccrueds map[common.Address]*big.Int
}

// NewPoolManager creates a new PoolManager instance.
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools:                make(map[libraries.PoolId]*libraries.Pool),
		currencyDelta:        libraries.NewCurrencyDelta(),
		protocolFeesAccrueds: make(map[common.Address]*big.Int),
	}
}

// Initialize creates and initializes a new Pool with the given PoolKey and starting price.
//
// Parameters:
// - key: PoolKey struct containing currency pair, fee, and tick spacing.
// - sqrtPriceX96: initial square root price of the pool, Q64.96 fixed-point.
//
// Returns:
//   - int: the initialized tick corresponding to sqrtPriceX96.
//   - error: any error encountered during initialization, including invalid tick spacing,
//     currencies out of order, or pool already existing.
//
// Errors:
// - ErrTickSpacingTooLarge if key.TickSpacing > utils.MaxTickSpacing.
// - ErrTickSpacingTooSmall if key.TickSpacing < utils.MinTickSpacing.
// - ErrCurrenciesOutOfOrder if key.Currency0 >= key.Currency1.
// - ErrPoolAlreadyInitialized if the pool has already been initialized.
func (pm *PoolManager) Initialize(key libraries.PoolKey, sqrtPriceX96 *big.Int) (int, error) {
	if key.TickSpacing > utils.MaxTickSpacing {
		return 0, ErrTickSpacingTooLarge
	}

	if key.TickSpacing < utils.MinTickSpacing {
		return 0, ErrTickSpacingTooSmall
	}

	if isSorted, err := key.Currency0.Lt(key.Currency1); err != nil || !isSorted {
		return 0, ErrCurrenciesOutOfOrder
	}

	id, err := key.ToId()
	if err != nil {
		return 0, err
	}

	if _, exists := pm.pools[id]; exists {
		return 0, ErrPoolAlreadyInitialized
	}

	lpFee, err := key.Fee.InitialValue()
	if err != nil {
		return 0, err
	}

	pool := libraries.NewPool()

	tick, err := pool.Initialize(sqrtPriceX96, libraries.LPFee(lpFee))
	if err != nil {
		return 0, err
	}

	pm.pools[id] = pool

	return tick, nil
}

// SetProcolFee updates the protocol fee configuration for a given pool.
//
// This function mirrors the governance-level protocol fee update in AMM systems
// such as Uniswap v4. The new protocol fee will affect subsequent swaps.
//
// Params:
//   - key: PoolKey uniquely identifying the pool
//   - newProtocolFee: new protocol fee configuration
//
// Returns:
//   - error if pool is not initialized or id conversion fails
//
// State Changes:
//   - Mutates pool protocol fee configuration
func (pm *PoolManager) SetProcolFee(key libraries.PoolKey, newProtocolFee libraries.ProtocolFee) error {
	id, err := key.ToId()
	if err != nil {
		return err
	}

	pool, ok := pm.pools[id]
	if !ok {
		return ErrPoolNotInitialized
	}

	return pool.SetProtocolFee(newProtocolFee)
}

// ModifyLiquidity adds or removes liquidity for a position owner.
//
// This function:
//  1. Updates position liquidity
//  2. Collects any accrued fees
//  3. Accounts net balance delta to the owner
//
// Returns:
//   - ownerDelta: net token balance change (principal + fees)
//   - feesAccrued: fees earned during position lifetime
//   - error if pool not initialized or liquidity update fails
//
// Accounting Model:
//
//	ownerDelta = principalDelta + feesAccrued
//	and is recorded via accountPoolBalanceDelta()
//
// This mirrors Uniswap v4's modifyLiquidity behavior.
func (pm *PoolManager) ModifyLiquidity(key libraries.PoolKey, params types.ModifyLiquidityParams, owner common.Address,
) (ownerDelta, feesAccrued libraries.BalanceDelta, err error) {
	id, err := key.ToId()
	if err != nil {
		return libraries.ZeroBalanceDelta, libraries.ZeroBalanceDelta, err
	}
	pool, ok := pm.pools[id]
	if !ok {
		return libraries.ZeroBalanceDelta, libraries.ZeroBalanceDelta, ErrPoolNotInitialized
	}

	principalDelta, feesAccrued, err := pool.ModifyLiquidity(
		libraries.ModifyLiquidityParams{
			Owner:          owner,
			TickLower:      params.TickLower,
			TickUpper:      params.TickUpper,
			LiquidityDelta: params.LiquidityDelta,
			TickSpacing:    key.TickSpacing,
			Salt:           params.Salt,
		},
	)
	if err != nil {
		return libraries.ZeroBalanceDelta, libraries.ZeroBalanceDelta, err
	}

	ownerDelta = principalDelta.Add(feesAccrued)

	pm.accountPoolBalanceDelta(key, ownerDelta, owner)

	return ownerDelta, feesAccrued, nil
}

// Swap executes a token swap against the specified pool.
//
// Behavior:
//   - Validates swap amount
//   - Ensures pool is initialized
//   - Executes core swap logic
//   - Accounts protocol fee
//   - Records user balance delta
//
// Params:
//   - key: pool identifier
//   - params: swap parameters (amount, direction, price limit)
//   - owner: swap initiator
//
// Returns:
//   - BalanceDelta: net token change for owner
//   - error if swap fails
//
// Note:
//
//	Protocol fee is always charged on input token.
func (pm *PoolManager) Swap(key libraries.PoolKey, params types.SwapParams, owner common.Address,
) (libraries.BalanceDelta, error) {

	if params.AmountSpecified.Sign() == 0 {
		return libraries.ZeroBalanceDelta, ErrSwapAmountCannotBeZero
	}

	id, err := key.ToId()
	if err != nil {
		return libraries.ZeroBalanceDelta, err
	}
	pool, ok := pm.pools[id]
	if !ok {
		return libraries.ZeroBalanceDelta, ErrPoolNotInitialized
	}

	if err := pool.CheckPoolInitialized(); err != nil {
		return libraries.ZeroBalanceDelta, err
	}

	inputCurrency := key.Currency0
	if !params.ZeroForOne {
		inputCurrency = key.Currency1
	}

	swapDelta, err := pm.swap(
		pool,
		libraries.SwapParams{
			AmountSpecified:   params.AmountSpecified,
			TickSpacing:       key.TickSpacing,
			ZeroForOne:        params.ZeroForOne,
			SqrtPriceLimitX96: params.SqrtPriceLimitX96,
		},
		inputCurrency.Address(),
	)
	if err != nil {
		return libraries.ZeroBalanceDelta, err
	}

	pm.accountPoolBalanceDelta(key, swapDelta, owner)

	return swapDelta, nil
}

// Donate transfers tokens directly into pool reserves without receiving liquidity.
//
// This increases pool reserves and benefits LPs proportionally.
//
// Returns error if:
//   - pool not initialized
//   - donation execution fails
//
// State Changes:
//   - pool reserves increase
//   - owner internal balance delta updated
func (pm *PoolManager) Donate(key libraries.PoolKey, amount0, amount1 *big.Int, owner common.Address) error {
	id, err := key.ToId()
	if err != nil {
		return err
	}
	pool, ok := pm.pools[id]
	if !ok {
		return ErrPoolNotInitialized
	}

	if err := pool.CheckPoolInitialized(); err != nil {
		return err
	}

	delta, err := pool.Donate(amount0, amount1)
	if err != nil {
		return err
	}

	pm.accountPoolBalanceDelta(key, delta, owner)

	return nil
}

// accountDelta applies a token balance delta to a user.
//
// This function updates internal accounting only and does not
// perform token transfers.
func (pm *PoolManager) accountDelta(currency common.Address, delta *big.Int, owner common.Address) {
	if delta.Sign() == 0 {
		return
	}
	pm.currencyDelta.ApplyDelta(owner, currency, delta)
}

// accountPoolBalanceDelta applies both token0 and token1 deltas.
func (pm *PoolManager) accountPoolBalanceDelta(key libraries.PoolKey, delta libraries.BalanceDelta, owner common.Address) {
	pm.accountDelta(key.Currency0.Address(), delta.Amount0, owner)
	pm.accountDelta(key.Currency1.Address(), delta.Amount1, owner)
}

// swap executes core pool swap logic and extracts protocol fees.
//
// Returns:
//   - BalanceDelta
//   - error
//
// If protocol fee is positive, it is accrued internally.
func (pm *PoolManager) swap(pool *libraries.Pool, params libraries.SwapParams, inputCurrency common.Address) (libraries.BalanceDelta, error) {
	delta, amountToProtocol, _, _, err := pool.Swap(params)
	if err != nil {
		return libraries.ZeroBalanceDelta, err
	}

	if amountToProtocol.Sign() > 0 {
		pm.updateProtocolFees(inputCurrency, amountToProtocol)
	}

	return delta, nil
}

// updateProtocolFees accumulates protocol fees per currency.
//
// Protocol fees are stored separately from pool reserves
// and can later be withdrawn by governance.
func (pm *PoolManager) updateProtocolFees(currency common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	if pm.protocolFeesAccrueds[currency] == nil {
		pm.protocolFeesAccrueds[currency] = big.NewInt(0)
	}
	pm.protocolFeesAccrueds[currency] = new(big.Int).Add(pm.protocolFeesAccrueds[currency], amount)
}

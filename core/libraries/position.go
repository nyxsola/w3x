package libraries

import (
	"math/big"

	"github.com/aicora/go-uniswap/core/utils"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrCannotUpdateEmptyPosition is returned when attempting to update a position
	// that has zero liquidity and the liquidity delta is also zero.
	ErrCannotUpdateEmptyPosition = errors.New("cannot update empty position")

	// ErrLiquidityUnderflow is returned when liquidity would become negative.
	ErrLiquidityUnderflow = errors.New("liquidity underflow")

	// ErrLiquidityOverflow is returned when liquidity would exceed uint128 max (2^128-1).
	ErrLiquidityOverflow = errors.New("liquidity overflow")
)

var (
	// maxUint128 is the maximum value of uint128: 2^128 - 1
	maxUint128 = new(big.Int).Sub(utils.Q128, big.NewInt(1))
)

// State represents a liquidity position between a lower and upper tick boundary.
//
// Liquidity is stored as uint128, and fee growths are tracked as uint256 per currency.
type State struct {
	Liquidity                *big.Int // uint128, amount of liquidity owned by this position
	FeeGrowthInside0LastX128 *big.Int // uint256, fee growth in currency0 at last update
	FeeGrowthInside1LastX128 *big.Int // uint256, fee growth in currency1 at last update
}

// Update updates the position with a liquidity delta and calculates fees owed.
//
// Parameters:
//   - liquidityDelta: the change in liquidity for this position (can be positive or negative)
//   - feeGrowthInside0X128: all-time fee growth in currency0 inside this position's tick boundaries
//   - feeGrowthInside1X128: all-time fee growth in currency1 inside this position's tick boundaries
//
// Returns:
//   - feesOwed0: amount of currency0 owed to the position owner
//   - feesOwed1: amount of currency1 owed to the position owner
//   - err: error if update is invalid, including empty position, underflow, or overflow
//
// Notes:
//   - Ensures liquidity never underflows (negative) or exceeds uint128 max.
//   - Fee calculations are done with high-precision MulDiv to emulate Solidity FullMath.mulDiv.
func (s *State) Update(liquidityDelta *big.Int, feeGrowthInside0X128, feeGrowthInside1X128 *big.Int) (feesOwed0, feesOwed1 *big.Int, err error) {
	if liquidityDelta.Sign() == 0 && s.Liquidity.Sign() == 0 {
		return nil, nil, ErrCannotUpdateEmptyPosition
	}

	// Update liquidity
	s.Liquidity = new(big.Int).Add(s.Liquidity, liquidityDelta)
	if s.Liquidity.Sign() < 0 {
		return nil, nil, ErrLiquidityUnderflow
	}
	if s.Liquidity.Cmp(maxUint128) > 0 {
		return nil, nil, ErrLiquidityOverflow
	}

	// Calculate fees owed in currency0
	delta0 := new(big.Int).Sub(feeGrowthInside0X128, s.FeeGrowthInside0LastX128)
	feesOwed0, err = utils.MulDiv(delta0, s.Liquidity, utils.Q128)
	if err != nil {
		return
	}

	// Calculate fees owed in currency1
	delta1 := new(big.Int).Sub(feeGrowthInside1X128, s.FeeGrowthInside1LastX128)
	feesOwed1, err = utils.MulDiv(delta1, s.Liquidity, utils.Q128)
	if err != nil {
		return
	}

	// Update last fee growths
	s.FeeGrowthInside0LastX128.Set(feeGrowthInside0X128)
	s.FeeGrowthInside1LastX128.Set(feeGrowthInside1X128)

	return feesOwed0, feesOwed1, nil
}

// PositionManager manages multiple liquidity positions indexed by position keys.
type PositionManager struct {
	states map[[32]byte]*State
}

// NewPositionManager creates a new PositionManager with initialized state map.
//
// The PositionManager holds all positions in memory. Each position is identified
// by a unique key generated from owner, tick boundaries, and salt.
func NewPositionManager() *PositionManager {
	return &PositionManager{
		states: make(map[[32]byte]*State),
	}
}

// Get returns the State of a position for a given owner and tick boundaries.
//
// Parameters:
//   - owner: the Ethereum address of the position owner
//   - tickLower: the lower tick boundary of the position
//   - tickUpper: the upper tick boundary of the position
//   - salt: unique value to differentiate multiple positions in the same tick range
//
// Returns:
//   - pointer to the State; if position does not exist, initializes a new State with zero values
//
// Notes:
//   - This method lazily initializes the State in the internal map if it does not exist.
//   - Position keys are generated using utils.ComputePositionKey, compatible with Solidity keccak256(abi.encodePacked(...)).
func (p *PositionManager) Get(owner common.Address, tickLower, tickUpper int, salt [32]byte) *State {
	key := utils.ComputePositionKey(owner, tickLower, tickUpper, salt)
	state, ok := p.states[key]
	if !ok {
		state = &State{
			Liquidity:                big.NewInt(0),
			FeeGrowthInside0LastX128: big.NewInt(0),
			FeeGrowthInside1LastX128: big.NewInt(0),
		}
		p.states[key] = state
	}
	return state
}

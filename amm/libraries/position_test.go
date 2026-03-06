package libraries

import (
	"math/big"
	"testing"

	"github.com/aicora/w3x/amm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestPositionManager_GetAndUpdate(t *testing.T) {
	pm := NewPositionManager()

	owner := common.HexToAddress("0x1234567890123456789012345678901234567890")
	tickLower := int(-100)
	tickUpper := int(100)
	var salt [32]byte
	copy(salt[:], []byte("salt-unique-value-1234567890"))

	state := pm.Get(owner, tickLower, tickUpper, salt)
	require.NotNil(t, state)
	require.Equal(t, 0, state.Liquidity.Sign())
	require.Equal(t, 0, state.FeeGrowthInside0LastX128.Sign())
	require.Equal(t, 0, state.FeeGrowthInside1LastX128.Sign())

	liquidityDelta := big.NewInt(1000)
	feeGrowth0X128 := new(big.Int).Set(utils.Q128)
	feeGrowth1X128 := new(big.Int).Set(utils.Q128)

	fees0, fees1, err := state.Update(liquidityDelta, feeGrowth0X128, feeGrowth1X128)
	require.NoError(t, err)
	require.Equal(t, int64(1000), state.Liquidity.Int64())
	// feesOwed0 = (feeGrowth - 0) * liquidity / Q128 = 1000 * Q128 / Q128 = 1000
	require.Equal(t, int64(1000), fees0.Int64())
	require.Equal(t, int64(1000), fees1.Int64())

	feeGrowth0X128.Add(feeGrowth0X128, utils.Q128)
	feeGrowth1X128.Add(feeGrowth1X128, utils.Q128)

	fees0, fees1, err = state.Update(big.NewInt(0), feeGrowth0X128, feeGrowth1X128)
	require.NoError(t, err)
	// feesOwed = liquidity * (feeGrowthDelta)/Q128 = 1000 * Q128 / Q128 = 1000
	require.Equal(t, int64(1000), fees0.Int64())
	require.Equal(t, int64(1000), fees1.Int64())

	stateEmpty := pm.Get(owner, 200, 300, salt)
	require.Equal(t, 0, stateEmpty.Liquidity.Sign())
	_, _, err = stateEmpty.Update(big.NewInt(0), utils.Q128, utils.Q128)
	require.ErrorIs(t, err, ErrCannotUpdateEmptyPosition)

	// liquidity overflow
	stateOverflow := pm.Get(owner, 400, 500, salt)
	stateOverflow.Liquidity = new(big.Int).Sub(maxUint128, big.NewInt(10))
	_, _, err = stateOverflow.Update(big.NewInt(20), utils.Q128, utils.Q128)
	require.ErrorIs(t, err, ErrLiquidityOverflow)

	// liquidity underflow
	stateUnderflow := pm.Get(owner, 600, 700, salt)
	stateUnderflow.Liquidity = big.NewInt(5)
	_, _, err = stateUnderflow.Update(big.NewInt(-10), utils.Q128, utils.Q128)
	require.ErrorIs(t, err, ErrLiquidityUnderflow)
}

func TestPositionManager_MultiplePositions(t *testing.T) {
	pm := NewPositionManager()
	owner1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	owner2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	var salt [32]byte

	pos1 := pm.Get(owner1, 0, 100, salt)
	pos2 := pm.Get(owner2, 0, 100, salt)
	pos1.Liquidity.SetInt64(1234)
	pos2.Liquidity.SetInt64(5678)
	require.Equal(t, int64(1234), pos1.Liquidity.Int64())
	require.Equal(t, int64(5678), pos2.Liquidity.Int64())
}

func TestCalculatePositionKeyConsistency(t *testing.T) {
	owner := common.HexToAddress("0x3333333333333333333333333333333333333333")
	tickLower := int(-50)
	tickUpper := int(50)
	var salt [32]byte
	copy(salt[:], []byte("salt-test"))

	key1 := utils.ComputePositionKey(owner, tickLower, tickUpper, salt)
	key2 := utils.ComputePositionKey(owner, tickLower, tickUpper, salt)
	require.Equal(t, key1, key2, "same inputs should produce same key")

	key3 := utils.ComputePositionKey(owner, tickLower, tickUpper+1, salt)
	require.NotEqual(t, key1, key3)
}

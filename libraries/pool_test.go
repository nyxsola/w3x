package libraries

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/aicora/go-uniswap/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	
	tick, err := p.Initialize(price, utils.LPFee(3000))
	assert.NoError(t, err)
	assert.Equal(t, 0, tick)

	assert.Equal(t, 0, p.slot0.Tick)
	assert.True(t, p.slot0.SqrtPriceX96.Cmp(price) == 0)
}

func TestInitializeTwice(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	_, err := p.Initialize(price, utils.LPFee(3000))
	assert.ErrorIs(t, err, ErrPoolAlreadyInitialized)
}

func TestDonate_NoLiquidity(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	_, err := p.Donate(big.NewInt(1000), big.NewInt(0))
	assert.ErrorIs(t, err, ErrNoLiquidity)
}

func TestModifyLiquidity_InRange(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:         common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:     -60,
		TickUpper:     60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:   60,
	}

	delta, feeDelta, err := p.ModifyLiquidity(params)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.NotNil(t, feeDelta)

	assert.True(t, p.liquidity.Sign() > 0)
}

func TestModifyLiquidity_Remove(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:         common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:     -60,
		TickUpper:     60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:   60,
	}

	_, _, _ = p.ModifyLiquidity(params)

	params.LiquidityDelta = big.NewInt(-1_000_000)

	_, _, err := p.ModifyLiquidity(params)
	assert.NoError(t, err)
	assert.Equal(t, 0, p.liquidity.Sign())
}

func TestDonate_WithLiquidity(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:         common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:     -60,
		TickUpper:     60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:   60,
	}

	_, _, _ = p.ModifyLiquidity(params)
	
	_, err := p.Donate(big.NewInt(1000), big.NewInt(0))
	assert.NoError(t, err)
	
	assert.True(t, p.feeGrowthGlobal0X128.Sign() > 0)
}

func TestSwapVariousScenarios(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, utils.LPFee(3000))

	delta, _, _ := p.ModifyLiquidity(ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -600,
		TickUpper:      600,
		LiquidityDelta: big.NewInt(1_000_000_000),
		TickSpacing:    60,
	})
	fmt.Println("delta", delta)
	tests := []struct {
		name             string
		zeroForOne       bool
		amountSpecified  int64
		percentOffset int64
	}{
		{"ZeroForOne_exactIn_small", true, 10000, -1},
		{"ZeroForOne_exactOut_small", true, -10000, -1},
		{"OneForZero_exactIn_small", false, 10000, 1},
		{"OneForZero_exactOut_small", false, -10000, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// price * percentOffset / 100
			delta := new(big.Int).Mul(price, big.NewInt(tt.percentOffset))
			delta.Div(delta, big.NewInt(100))

			// limit = price + delta
			limit := new(big.Int).Add(price, delta)

			params := SwapParams{
				ZeroForOne:        tt.zeroForOne,
				AmountSpecified:   big.NewInt(tt.amountSpecified),
				SqrtPriceLimitX96: limit,
				TickSpacing:       60,
			}

			swapDelta, amountToProtocol, swapfee, result, err := p.Swap(params)
			assert.NoError(t, err)
			assert.NotNil(t, swapDelta)
			assert.NotNil(t, amountToProtocol)
			assert.NotNil(t, result.SqrtPriceX96)

			if tt.zeroForOne {
				assert.True(t, result.SqrtPriceX96.Cmp(price) < 0)
			} else {
				assert.True(t, result.SqrtPriceX96.Cmp(price) > 0)
			}

			fmt.Println(tt.name, swapDelta, swapfee, result.SqrtPriceX96)
			price.Set(result.SqrtPriceX96)
		})
	}
}
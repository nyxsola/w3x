package libraries

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/aicora/w3x/amm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)

	tick, err := p.Initialize(price, LPFee(3000))
	assert.NoError(t, err)
	assert.Equal(t, 0, tick)

	assert.Equal(t, 0, p.slot0.Tick)
	assert.True(t, p.slot0.SqrtPriceX96.Cmp(price) == 0)
}

func TestInitializeTwice(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, LPFee(3000))

	_, err := p.Initialize(price, LPFee(3000))
	assert.ErrorIs(t, err, ErrPoolAlreadyInitialized)
}

func TestPool_SetProtocolFee(t *testing.T) {
	p := NewPool()

	err := p.SetProtocolFee(ProtocolFee(500))
	if err == nil {
		t.Fatal("expected error when pool not initialized")
	}

	price, _ := utils.GetSqrtPriceAtTick(0)

	p.Initialize(price, 0)

	newFee := ProtocolFee(300)
	err = p.SetProtocolFee(newFee)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.slot0.ProtocolFee != newFee {
		t.Fatalf("protocol fee not updated")
	}
}

func TestPool_SetLPFee(t *testing.T) {
	p := NewPool()

	err := p.SetLPFee(LPFee(3000))
	if err == nil {
		t.Fatal("expected error when pool not initialized")
	}

	price, _ := utils.GetSqrtPriceAtTick(0)

	p.Initialize(price, 0)

	newFee := LPFee(2500)
	err = p.SetLPFee(newFee)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.slot0.LPFee != newFee {
		t.Fatalf("lp fee not updated")
	}
}

func TestPool_ClearTick(t *testing.T) {
	p := NewPool()
	price, _ := utils.GetSqrtPriceAtTick(0)
	p.Initialize(price, 0)

	testTick := 100

	info := p.tickManager.Get(testTick)

	info.LiquidityNet.Add(info.LiquidityNet, big.NewInt(100))

	p.ClearTick(testTick)

	info = p.tickManager.Get(testTick)
	if info.LiquidityNet.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected LiquidityNet 0, got %s", info.LiquidityNet.String())
	}
}

func TestPool_CrossTick_Basic(t *testing.T) {
	p := NewPool()
	price, _ := utils.GetSqrtPriceAtTick(0)
	p.Initialize(price, 0)

	tick := 100

	initialOutside0 := big.NewInt(1000)
	initialOutside1 := big.NewInt(2000)

	info := p.tickManager.Get(tick)
	info.FeeGrowthOutside0X128 = new(big.Int).Set(initialOutside0)
	info.FeeGrowthOutside1X128 = new(big.Int).Set(initialOutside1)
	info.LiquidityNet = big.NewInt(500)

	feeGlobal0 := big.NewInt(5000)
	feeGlobal1 := big.NewInt(8000)

	liquidityNet := p.CrossTick(tick, feeGlobal0, feeGlobal1)

	updated := p.tickManager.Get(tick)

	expected0 := new(big.Int).Sub(feeGlobal0, initialOutside0)
	expected1 := new(big.Int).Sub(feeGlobal1, initialOutside1)

	if updated.FeeGrowthOutside0X128.Cmp(expected0) != 0 {
		t.Fatalf("feeGrowthOutside0 incorrect")
	}

	if updated.FeeGrowthOutside1X128.Cmp(expected1) != 0 {
		t.Fatalf("feeGrowthOutside1 incorrect")
	}

	if liquidityNet.Cmp(big.NewInt(500)) != 0 {
		t.Fatalf("liquidityNet incorrect")
	}
}

func TestDonate_NoLiquidity(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, LPFee(3000))

	_, err := p.Donate(big.NewInt(1000), big.NewInt(0))
	assert.ErrorIs(t, err, ErrNoLiquidity)
}

func TestModifyLiquidity_InRange(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -60,
		TickUpper:      60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:    60,
	}

	delta, feeDelta, err := p.ModifyLiquidity(params)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.NotNil(t, feeDelta)

	assert.True(t, p.liquidity.Sign() > 0)
}

func TestModifyLiquidity_OutOfRange1(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(-70)
	_, _ = p.Initialize(price, LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -60,
		TickUpper:      60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:    60,
	}

	delta, feeDelta, err := p.ModifyLiquidity(params)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.NotNil(t, feeDelta)

	assert.True(t, p.liquidity.Sign() == 0)
}

func TestModifyLiquidity_OutOfRange2(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(70)
	_, _ = p.Initialize(price, LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -60,
		TickUpper:      60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:    60,
	}

	delta, feeDelta, err := p.ModifyLiquidity(params)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.NotNil(t, feeDelta)

	assert.True(t, p.liquidity.Sign() == 0)
}

func TestModifyLiquidity_Remove(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(0)
	_, _ = p.Initialize(price, LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -60,
		TickUpper:      60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:    60,
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
	_, _ = p.Initialize(price, LPFee(3000))

	params := ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -60,
		TickUpper:      60,
		LiquidityDelta: big.NewInt(1_000_000),
		TickSpacing:    60,
	}

	_, _, _ = p.ModifyLiquidity(params)

	_, err := p.Donate(big.NewInt(1000), big.NewInt(1000))
	assert.NoError(t, err)

	assert.True(t, p.feeGrowthGlobal0X128.Sign() > 0)
}

func TestSwapVariousScenarios(t *testing.T) {
	p := NewPool()

	price, _ := utils.GetSqrtPriceAtTick(300)
	_, _ = p.Initialize(price, LPFee(3000))
	pfee, _ := NewProtocolFee(10, 10)
	p.SetProtocolFee(pfee)
	delta, _, _ := p.ModifyLiquidity(ModifyLiquidityParams{
		Owner:          common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567"),
		TickLower:      -600,
		TickUpper:      600,
		LiquidityDelta: big.NewInt(1_000_000_000),
		TickSpacing:    60,
	})
	fmt.Println("delta", delta)
	tests := []struct {
		name            string
		zeroForOne      bool
		amountSpecified int64
		percentOffset   int64
	}{
		{"ZeroForOne_exactIn_small", true, -10000, -1},
		{"ZeroForOne_exactOut_small", true, 10000, -1},
		{"OneForZero_exactIn_small", false, -10000, 1},
		{"OneForZero_exactOut_small", false, 10000, 1},
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

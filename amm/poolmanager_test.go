package core

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/aicora/w3x/amm/libraries"
	"github.com/aicora/w3x/amm/types"
	"github.com/aicora/w3x/amm/utils"
)

func mockPoolKey() libraries.PoolKey {
	lpfee, _ := libraries.NewFee(3000)
	return libraries.PoolKey{
		Currency0:   libraries.NewCurrency(1, common.HexToAddress("0x0000000000000000000000000000000000000001"), 18, "A", "TokenA"),
		Currency1:   libraries.NewCurrency(1, common.HexToAddress("0x0000000000000000000000000000000000000002"), 6, "B", "TokenB"),
		Fee:         lpfee,
		TickSpacing: 60,
	}
}

func TestInitialize_Success(t *testing.T) {
	pm := NewPoolManager()
	key := mockPoolKey()

	sqrtPriceX96 := new(big.Int).Mul(big.NewInt(1), utils.Q96)

	tick, err := pm.Initialize(key, sqrtPriceX96)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tick != 0 {
		t.Fatalf("expected zero tick after initialization")
	}
}

func TestInitialize_Duplicate(t *testing.T) {
	pm := NewPoolManager()
	key := mockPoolKey()

	sqrtPriceX96 := new(big.Int).Mul(big.NewInt(1), utils.Q96)

	_, _ = pm.Initialize(key, sqrtPriceX96)
	_, err := pm.Initialize(key, sqrtPriceX96)

	if err != ErrPoolAlreadyInitialized {
		t.Fatalf("expected ErrPoolAlreadyInitialized, got %v", err)
	}
}

func TestSwap_ZeroAmount(t *testing.T) {
	pm := NewPoolManager()
	key := mockPoolKey()

	_, err := pm.Swap(
		key,
		types.SwapParams{
			AmountSpecified: big.NewInt(0),
		},
		common.HexToAddress("0xabc"),
	)

	if err != ErrSwapAmountCannotBeZero {
		t.Fatalf("expected ErrSwapAmountCannotBeZero, got %v", err)
	}
}

func TestDonate_NotInitialized(t *testing.T) {
	pm := NewPoolManager()
	key := mockPoolKey()

	err := pm.Donate(
		key,
		big.NewInt(100),
		big.NewInt(100),
		common.HexToAddress("0x0000000000000000000000000000000000000abc"),
	)

	if err != ErrPoolNotInitialized {
		t.Fatalf("expected ErrPoolNotInitialized, got %v", err)
	}
}

func TestProtocolFeeAccumulation(t *testing.T) {
	pm := NewPoolManager()

	currency := common.HexToAddress("0x0000000000000000000000000000000000000001")
	amount := big.NewInt(1000)

	pm.updateProtocolFees(currency, amount)

	stored := pm.protocolFeesAccrueds[currency]
	if stored == nil {
		t.Fatalf("expected protocol fee stored")
	}

	if stored.Cmp(amount) != 0 {
		t.Fatalf("expected %v, got %v", amount, stored)
	}
}

func TestAccountDelta_NoPanicOnZero(t *testing.T) {
	pm := NewPoolManager()

	pm.accountDelta(
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		big.NewInt(0),
		common.HexToAddress("0x000000000000000000000000000000000000user"),
	)
}

func TestModifyLiquidity_NotInitialized(t *testing.T) {
	pm := NewPoolManager()
	key := mockPoolKey()

	_, _, err := pm.ModifyLiquidity(
		key,
		types.ModifyLiquidityParams{},
		common.HexToAddress("0x000000000000000000000000000000000000user"),
	)

	if err != ErrPoolNotInitialized {
		t.Fatalf("expected ErrPoolNotInitialized, got %v", err)
	}
}

func TestPoolManager_SwapIntegration(t *testing.T) {
	pm := NewPoolManager()

	owner := common.HexToAddress("0x00000000000000000000000000000000000000aa")

	key := libraries.PoolKey{
		Currency0:   libraries.NewCurrency(1, common.HexToAddress("0x0000000000000000000000000000000000000001"), 18, "A", "TokenA"),
		Currency1:   libraries.NewCurrency(1, common.HexToAddress("0x0000000000000000000000000000000000000002"), 6, "B", "TokenB"),
		Fee:         libraries.LPFee(3000),
		TickSpacing: 60,
	}

	userDelta0 := pm.currencyDelta.GetDelta(owner, key.Currency0.Address())
	userDelta1 := pm.currencyDelta.GetDelta(owner, key.Currency1.Address())
	fmt.Println("userDelta", userDelta0, userDelta1)

	sqrtPriceX96 := new(big.Int).Mul(big.NewInt(1), utils.Q96)

	_, err := pm.Initialize(key, sqrtPriceX96)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	userDelta0 = pm.currencyDelta.GetDelta(owner, key.Currency0.Address())
	userDelta1 = pm.currencyDelta.GetDelta(owner, key.Currency1.Address())
	fmt.Println("userDelta", userDelta0, userDelta1)

	liquidity := big.NewInt(1_000_000)

	_, _, err = pm.ModifyLiquidity(
		key,
		types.ModifyLiquidityParams{
			TickLower:      -60,
			TickUpper:      60,
			LiquidityDelta: liquidity,
		},
		owner,
	)
	if err != nil {
		t.Fatalf("add liquidity failed: %v", err)
	}

	userDelta0 = pm.currencyDelta.GetDelta(owner, key.Currency0.Address())
	userDelta1 = pm.currencyDelta.GetDelta(owner, key.Currency1.Address())
	fmt.Println("userDelta", userDelta0, userDelta1)

	swapAmount := big.NewInt(-1000)

	delta, err := pm.Swap(
		key,
		types.SwapParams{
			ZeroForOne:        true,
			AmountSpecified:   swapAmount,
			SqrtPriceLimitX96: new(big.Int).Add(utils.MinSqrtPrice, big.NewInt(1)),
		},
		owner,
	)
	if err != nil {
		t.Fatalf("swap failed: %v", err)
	}

	if delta.Amount0.Sign() == 0 && delta.Amount1.Sign() == 0 {
		t.Fatalf("swap delta should not be zero")
	}

	userDelta0 = pm.currencyDelta.GetDelta(owner, key.Currency0.Address())
	userDelta1 = pm.currencyDelta.GetDelta(owner, key.Currency1.Address())
	fmt.Println("userDelta", userDelta0, userDelta1)

	if userDelta0 == nil {
		t.Fatalf("user delta should exist")
	}
}

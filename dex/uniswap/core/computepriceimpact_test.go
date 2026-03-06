package uniswapsdkcore

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func ethToken() *Currency {
	return NewCurrency(1, common.HexToAddress("0xETH0000000000000000000000000000000000"), 18, "ETH", "Ether")
}

func usdcToken() *Currency {
	return NewCurrency(1, common.HexToAddress("0xUSDC000000000000000000000000000000000"), 6, "USDC", "US Dollar Coin")
}

func TestComputePriceImpact(t *testing.T) {
	ETH := ethToken()
	USDC := usdcToken()

	// midPrice ETH/USDC = 2000
	midPrice := NewPrice(ETH, USDC, big.NewInt(1e18), big.NewInt(2000e6))

	// input: 1 ETH
	input := FromRawAmount(ETH, big.NewInt(1e18))

	// actual output: 1990 USDC
	output := FromRawAmount(USDC, big.NewInt(1990e6))

	impact, err := ComputePriceImpact(midPrice, input, output)
	if err != nil {
		t.Fatal(err)
	}

	if impact.ToFixed(4) != "0.5000" {
		t.Fatal("expected 0.5000%")
	}

}

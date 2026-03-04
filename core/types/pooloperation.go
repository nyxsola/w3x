package types

import "math/big"

// ModifyLiquidityParams contains parameters for modifying a liquidity position in a pool.
//
// It is the Go equivalent of Solidity's `ModifyLiquidityParams` struct.
// Fields:
//   - TickLower: the lower tick of the liquidity position
//   - TickUpper: the upper tick of the liquidity position
//   - LiquidityDelta: the amount of liquidity to add (positive) or remove (negative)
//   - Salt: optional value to differentiate multiple positions at the same tick range
type ModifyLiquidityParams struct {
	TickLower      int // int24 in Solidity, mapped to int32 in Go
	TickUpper      int
	LiquidityDelta *big.Int
	Salt           [32]byte
}

// SwapParams contains parameters for executing a swap in a pool.
//
// Go equivalent of Solidity's `SwapParams` struct.
// Fields:
//   - ZeroForOne: true if swapping token0 for token1; false otherwise
//   - AmountSpecified: desired input (negative for exactIn) or output (positive for exactOut)
//   - SqrtPriceLimitX96: sqrt price at which the swap stops, in Q64.96 fixed point format
type SwapParams struct {
	ZeroForOne        bool
	AmountSpecified   *big.Int
	SqrtPriceLimitX96 *big.Int // uint160 in Solidity, mapped to big.Int
}

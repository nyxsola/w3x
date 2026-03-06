package utils

import (
	"math/big"
)

// AddDelta applies a signed delta `y` to value `x` and returns the result.
//
// If y >= 0:
//
//	result = x + y
//
// If y < 0:
//
//	result = x - |y|
//
// This function does NOT mutate x or y.
// A new big.Int instance is always returned.
//
// Equivalent to Solidity pattern:
//
//	x = x + int256(delta)
//
// Commonly used in liquidity or balance adjustments where
// delta may be positive or negative.
func AddDelta(x, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}

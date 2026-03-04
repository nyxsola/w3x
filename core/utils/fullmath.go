package utils

import (
	"math/big"

	"github.com/pkg/errors"
)

var (
	// ErrZeroDenominator is returned when a division operation is attempted with a zero denominator.
	ErrZeroDenominator = errors.New("denominator is zero")
)

// MulDiv calculates floor(a * b / denominator) with arbitrary precision using big.Int.
//
// This function performs the multiplication and division in full precision and
// returns an error if the denominator is zero.
//
// Parameters:
//   - a: multiplicand
//   - b: multiplier
//   - denominator: divisor
//
// Returns:
//   - *big.Int: floor(a * b / denominator)
//   - error: if denominator == 0
func MulDiv(a, b, denominator *big.Int) (*big.Int, error) {
	if denominator.Sign() == 0 {
		return nil, ErrZeroDenominator
	}

	prod := new(big.Int).Mul(a, b) // a * b
	res := new(big.Int).Div(prod, denominator)

	return res, nil
}

// MulDivRoundingUp calculates ceil(a * b / denominator) with arbitrary precision using big.Int.
//
// If the division is not exact, the result is rounded up by 1.
// Returns an error if denominator is zero.
//
// Parameters:
//   - a: multiplicand
//   - b: multiplier
//   - denominator: divisor
//
// Returns:
//   - *big.Int: ceil(a * b / denominator)
//   - error: if denominator == 0
func MulDivRoundingUp(a, b, denominator *big.Int) (*big.Int, error) {
	if denominator.Sign() == 0 {
		return nil, ErrZeroDenominator
	}

	// Multiply first
	prod := new(big.Int).Mul(a, b)

	// Divide to get floor
	res := new(big.Int).Div(prod, denominator)

	// Check remainder to round up
	mod := new(big.Int).Mod(prod, denominator)
	if mod.Sign() > 0 {
		res.Add(res, big.NewInt(1))
	}

	return res, nil
}

// MulMod calculates (a * b) % m using big.Int arithmetic.
//
// Parameters:
//   - a: multiplicand
//   - b: multiplier
//   - m: modulus
//
// Returns:
//   - *big.Int: result of (a * b) mod m
func MulMod(a, b, m *big.Int) *big.Int {
	res := new(big.Int).Mul(a, b) // a * b
	return res.Mod(res, m)        // (a * b) % m
}

// AbsBigInt returns the absolute value of a big.Int.
//
// Parameters:
//   - x: the input *big.Int
//
// Returns:
//   - *big.Int: absolute value of x
func AbsBigInt(x *big.Int) *big.Int {
	if x.Sign() < 0 {
		return new(big.Int).Neg(x)
	}
	return new(big.Int).Set(x)
}

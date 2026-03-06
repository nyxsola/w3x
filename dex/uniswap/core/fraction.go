package uniswapsdkcore

import (
	"fmt"
	"math"
	"math/big"

	"github.com/shopspring/decimal"
)

// Fraction represents a rational number with a numerator and a denominator.
// All arithmetic operations are performed with arbitrary precision using big.Int.
type Fraction struct {
	Numerator   *big.Int // numerator of the fraction
	Denominator *big.Int // denominator of the fraction (must not be zero)
}

// NewFraction creates a new Fraction with the given numerator and denominator.
//
// Example:
//
//	f := NewFraction(big.NewInt(3), big.NewInt(4)) // 3/4
func NewFraction(numerator, denominator *big.Int) *Fraction {
	return &Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

// Quotient returns the integer part of the fraction (floor division).
//
// Example:
//
//	NewFraction(7, 3).Quotient() // returns 2
func (f *Fraction) Quotient() *big.Int {
	return new(big.Int).Div(f.Numerator, f.Denominator)
}

// Remainder returns the remainder of the fraction after floor division as a new Fraction.
//
// Example:
//
//	NewFraction(7, 3).Remainder() // returns 1/3
func (f *Fraction) Remainder() *Fraction {
	return NewFraction(new(big.Int).Rem(f.Numerator, f.Denominator), f.Denominator)
}

// Invert returns the reciprocal of the fraction.
//
// Example:
//
//	NewFraction(3, 4).Invert() // returns 4/3
func (f *Fraction) Invert() *Fraction {
	return NewFraction(f.Denominator, f.Numerator)
}

// Add returns a new Fraction representing f + other.
//
// Example:
//
//	NewFraction(1,2).Add(NewFraction(1,3)) // returns 5/6
func (f *Fraction) Add(other *Fraction) *Fraction {
	if f.Denominator.Cmp(other.Denominator) == 0 {
		return NewFraction(new(big.Int).Add(f.Numerator, other.Numerator), f.Denominator)
	}
	return NewFraction(
		new(big.Int).Add(new(big.Int).Mul(f.Numerator, other.Denominator), new(big.Int).Mul(other.Numerator, f.Denominator)),
		new(big.Int).Mul(f.Denominator, other.Denominator))
}

// Subtract returns a new Fraction representing f - other.
//
// Example:
//
//	NewFraction(3,4).Subtract(NewFraction(1,2)) // returns 1/4
func (f *Fraction) Subtract(other *Fraction) *Fraction {
	if f.Denominator.Cmp(other.Denominator) == 0 {
		return NewFraction(new(big.Int).Sub(f.Numerator, other.Numerator), f.Denominator)
	}
	return NewFraction(
		new(big.Int).Sub(new(big.Int).Mul(f.Numerator, other.Denominator), new(big.Int).Mul(other.Numerator, f.Denominator)),
		new(big.Int).Mul(f.Denominator, other.Denominator))
}

// Multiply returns a new Fraction representing f * other.
//
// Example:
//
//	NewFraction(2,3).Multiply(NewFraction(3,4)) // returns 1/2
func (f *Fraction) Multiply(other *Fraction) *Fraction {
	return NewFraction(new(big.Int).Mul(f.Numerator, other.Numerator), new(big.Int).Mul(f.Denominator, other.Denominator))
}

// Divide returns a new Fraction representing f / other.
//
// Example:
//
//	NewFraction(2,3).Divide(NewFraction(3,4)) // returns 8/9
func (f *Fraction) Divide(other *Fraction) *Fraction {
	return NewFraction(new(big.Int).Mul(f.Numerator, other.Denominator), new(big.Int).Mul(f.Denominator, other.Numerator))
}

// LessThan returns true if f < other.
//
// Example:
//
//	NewFraction(1,2).LessThan(NewFraction(2,3)) // true
func (f *Fraction) LessThan(other *Fraction) bool {
	return new(big.Int).Mul(f.Numerator, other.Denominator).Cmp(new(big.Int).Mul(other.Numerator, f.Denominator)) < 0
}

// EqualTo returns true if f == other.
//
// Example:
//
//	NewFraction(1,2).EqualTo(NewFraction(2,4)) // true
func (f *Fraction) EqualTo(other *Fraction) bool {
	return new(big.Int).Mul(f.Numerator, other.Denominator).Cmp(new(big.Int).Mul(other.Numerator, f.Denominator)) == 0
}

// GreaterThan returns true if f > other.
//
// Example:
//
//	NewFraction(3,4).GreaterThan(NewFraction(2,3)) // true
func (f *Fraction) GreaterThan(other *Fraction) bool {
	return new(big.Int).Mul(f.Numerator, other.Denominator).Cmp(new(big.Int).Mul(other.Numerator, f.Denominator)) > 0
}

// ToSignificant returns a string representing the fraction rounded to the specified number of significant digits.
//
// Example:
//
//	NewFraction(big.NewInt(125), big.NewInt(1)).ToSignificant(2) // "130"
func (f *Fraction) ToSignificant(significantDigits int32) string {
	return roundToSignificantFigures(f, significantDigits).String()
}

// ToFixed returns a string representing the fraction rounded to a fixed number of decimal places.
//
// Example:
//
//	NewFraction(big.NewInt(1), big.NewInt(3)).ToFixed(2) // "0.33"
func (f *Fraction) ToFixed(decimalPlaces int32) string {
	return decimal.NewFromBigInt(f.Numerator, 0).Div(decimal.NewFromBigInt(f.Denominator, 0)).StringFixed(decimalPlaces)
}

var (
	oneInt = big.NewInt(1)
	twoInt = big.NewInt(2)
	tenInt = big.NewInt(10)
)

// roundToSignificantFigures returns a decimal representing the fraction rounded to a specified number of significant figures.
// It normalizes the result and removes trailing zeros.
//
// Notes:
// - If figures <= 0, returns 0
// - Uses arbitrary precision arithmetic via big.Int and shopspring/decimal
func roundToSignificantFigures(f *Fraction, figures int32) decimal.Decimal {
	if figures <= 0 {
		return decimal.Zero
	}
	d := decimal.NewFromBigInt(f.Numerator, 0).Div(decimal.NewFromBigInt(f.Denominator, 0))
	twoMant := d.Mul(decimal.NewFromFloat(math.Pow10(decimal.DivisionPrecision))).BigInt()
	twoMant.Abs(twoMant)
	twoMant.Mul(twoMant, twoInt)
	upper := big.NewInt(int64(figures))
	upper.Exp(tenInt, upper, nil)
	upper.Mul(upper, twoInt)
	upper.Sub(upper, oneInt)
	m := int64(0)
	for twoMant.Cmp(upper) >= 0 {
		upper.Mul(upper, tenInt)
		m++
	}
	if int64(d.Exponent())+m > int64(math.MaxInt32) {
		panic(fmt.Sprintf("exponent %d overflows an int32", int64(d.Exponent())+m))
	}
	return d.Round(-d.Exponent() - int32(m))
}

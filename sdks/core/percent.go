package sdkcore

import "math/big"

// OneHundred represents the fraction 100/1 and is used for percent calculations.
var OneHundred = NewFraction(big.NewInt(100), big.NewInt(1))

// Percent represents a value as a percentage.
//
// Percent wraps a Fraction, allowing arithmetic operations while keeping
// results in percentage format (0–100).
type Percent struct {
	*Fraction
}

// toPercent converts a Fraction to a Percent.
//
// Example:
//
//	f := NewFraction(big.NewInt(1), big.NewInt(2)) // 1/2
//	p := toPercent(f) // 50%
func toPercent(fraction *Fraction) *Percent {
	return NewPercent(fraction.Numerator, fraction.Denominator)
}

// NewPercent creates a new Percent from a numerator and denominator.
//
// Example:
//
//	p := NewPercent(big.NewInt(3), big.NewInt(4)) // 75%
func NewPercent(numerator, denominator *big.Int) *Percent {
	return &Percent{NewFraction(numerator, denominator)}
}

// Add returns a new Percent representing the sum of p and other.
//
// Example:
//
//	p1 := NewPercent(big.NewInt(25), big.NewInt(1)) // 25%
//	p2 := NewPercent(big.NewInt(30), big.NewInt(1)) // 30%
//	sum := p1.Add(p2) // 55%
func (p *Percent) Add(other *Percent) *Percent {
	return toPercent(p.Fraction.Add(other.Fraction))
}

// Subtract returns a new Percent representing the difference of p and other.
//
// Example:
//
//	p1 := NewPercent(big.NewInt(60), big.NewInt(1)) // 60%
//	p2 := NewPercent(big.NewInt(20), big.NewInt(1)) // 20%
//	diff := p1.Subtract(p2) // 40%
func (p *Percent) Subtract(other *Percent) *Percent {
	return toPercent(p.Fraction.Subtract(other.Fraction))
}

// Multiply returns a new Percent representing the product of p and other.
//
// Example:
//
//	p1 := NewPercent(big.NewInt(50), big.NewInt(1)) // 50%
//	p2 := NewPercent(big.NewInt(50), big.NewInt(1)) // 50%
//	product := p1.Multiply(p2) // 25%
func (p *Percent) Multiply(other *Percent) *Percent {
	return toPercent(p.Fraction.Multiply(other.Fraction))
}

// Divide returns a new Percent representing p divided by other.
//
// Example:
//
//	p1 := NewPercent(big.NewInt(50), big.NewInt(1)) // 50%
//	p2 := NewPercent(big.NewInt(25), big.NewInt(1)) // 25%
//	quotient := p1.Divide(p2) // 200%
func (p *Percent) Divide(other *Percent) *Percent {
	return toPercent(p.Fraction.Divide(other.Fraction))
}

// ToSignificant returns a string representation of the Percent with the specified
// number of significant digits.
//
// Example:
//
//	p := NewPercent(big.NewInt(1), big.NewInt(2)) // 50%
//	fmt.Println(p.ToSignificant(2)) // "50"
func (p *Percent) ToSignificant(significantDigits int32) string {
	return p.Fraction.Multiply(OneHundred).ToSignificant(significantDigits)
}

// ToFixed returns a string representation of the Percent with a fixed number of
// decimal places.
//
// Example:
//
//	p := NewPercent(big.NewInt(1), big.NewInt(3)) // ~33.33%
//	fmt.Println(p.ToFixed(2)) // "33.33"
func (p *Percent) ToFixed(decimalPlaces int32) string {
	return p.Fraction.Multiply(OneHundred).ToFixed(decimalPlaces)
}

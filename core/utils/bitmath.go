package utils

import (
	"errors"
	"math/big"
	"math/bits"
)

var ErrZeroInput = errors.New("input is zero")

// MostSignificantBit returns the index of the most significant bit set to 1
// in x. x is assumed to be non-zero.
func MostSignificantBit(x *big.Int) (int, error) {
	if x.Sign() == 0 {
		return 0, ErrZeroInput
	}

	words := x.Bits()
	wordSize := bits.UintSize

	for i := len(words) - 1; i >= 0; i-- {
		w := words[i]
		if w != 0 {
			msb := bits.Len(uint(w)) - 1
			return i*wordSize + msb, nil
		}
	}

	return 0, ErrZeroInput
}

// LeastSignificantBit returns the index of the least significant bit set to 1
// in x. x is assumed to be non-zero.
// For example, LSB(0b101000) = 3
func LeastSignificantBit(x *big.Int) (int, error) {
	if x.Sign() == 0 {
		return 0, ErrZeroInput
	}

	words := x.Bits() // []big.Word, 32/64-bit words depending on architecture
	wordSize := bits.UintSize

	for wordIndex, w := range words {
		if w != 0 {
			lsb := bits.TrailingZeros(uint(w))
			return wordIndex*wordSize + lsb, nil
		}
	}
	return 0, ErrZeroInput
}

package utils

import (
	"math/big"
	"testing"
)

func TestMostSignificantBit(t *testing.T) {
	tests := []struct {
		x        string
		expected int
		err      bool
	}{
		{"0", 0, true},
		{"1", 0, false},
		{"2", 1, false},
		{"8", 3, false},
		{"255", 7, false},
		{"1024", 10, false},
		{"18446744073709551615", 63, false}, // max uint64
		{"340282366920938463463374607431768211455", 127, false}, // max uint128
	}

	for _, tt := range tests {
		x, _ := new(big.Int).SetString(tt.x, 10)
		msb, err := MostSignificantBit(x)
		if (err != nil) != tt.err {
			t.Errorf("MostSignificantBit(%s) unexpected error: %v", tt.x, err)
			continue
		}
		if !tt.err && msb != tt.expected {
			t.Errorf("MostSignificantBit(%s) = %d, want %d", tt.x, msb, tt.expected)
		}
	}
}

func TestLeastSignificantBit(t *testing.T) {
	tests := []struct {
		x        string
		expected int
		err      bool
	}{
		{"0", 0, true},
		{"1", 0, false},
		{"2", 1, false},
		{"8", 3, false},
		{"255", 0, false},
		{"1024", 10, false},
		{"18446744073709551615", 0, false}, // max uint64
		{"340282366920938463463374607431768211455", 0, false}, // max uint128
	}

	for _, tt := range tests {
		x, _ := new(big.Int).SetString(tt.x, 10)
		lsb, err := LeastSignificantBit(x)
		if (err != nil) != tt.err {
			t.Errorf("LeastSignificantBit(%s) unexpected error: %v", tt.x, err)
			continue
		}
		if !tt.err && lsb != tt.expected {
			t.Errorf("LeastSignificantBit(%s) = %d, want %d", tt.x, lsb, tt.expected)
		}
	}
}

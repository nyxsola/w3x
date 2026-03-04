package utils

import (
	"math/big"
	"testing"
)

func TestAddDelta(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		expected string
	}{
		{"positive + positive", "10", "5", "15"},
		{"positive + negative", "10", "-3", "7"},
		{"negative + positive", "-10", "3", "-7"},
		{"negative + negative", "-10", "-5", "-15"},
		{"zero + positive", "0", "8", "8"},
		{"zero + negative", "0", "-8", "-8"},
		{"large numbers",
			"340282366920938463463374607431768211455", // 2^128-1
			"1",
			"340282366920938463463374607431768211456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, _ := new(big.Int).SetString(tt.x, 10)
			y, _ := new(big.Int).SetString(tt.y, 10)

			xCopy := new(big.Int).Set(x)
			yCopy := new(big.Int).Set(y)

			result := AddDelta(x, y)

			expected, _ := new(big.Int).SetString(tt.expected, 10)

			if result.Cmp(expected) != 0 {
				t.Errorf("AddDelta(%s, %s) = %s, want %s",
					tt.x, tt.y, result.String(), tt.expected)
			}

			if x.Cmp(xCopy) != 0 {
				t.Errorf("x was modified: got %s, want %s",
					x.String(), xCopy.String())
			}
			if y.Cmp(yCopy) != 0 {
				t.Errorf("y was modified: got %s, want %s",
					y.String(), yCopy.String())
			}
		})
	}
}

package uniswapsdkv4

import (
	"math/big"
	"testing"
)

func TestToHex(t *testing.T) {
	tests := []struct {
		name  string
		input *big.Int
		want  string
	}{
		{
			name:  "nil input",
			input: nil,
			want:  "0x00",
		},
		{
			name:  "zero value",
			input: big.NewInt(0),
			want:  "0x00",
		},
		{
			name:  "single byte value",
			input: big.NewInt(15),
			want:  "0x0f",
		},
		{
			name:  "multi byte value",
			input: big.NewInt(255),
			want:  "0xff",
		},
		{
			name:  "large value",
			input: new(big.Int).SetBytes([]byte{0x01, 0x23, 0x45}),
			want:  "0x012345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHex(tt.input)
			if got != tt.want {
				t.Errorf("ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMethodParametersInitialization(t *testing.T) {
	calldata := []byte{0xde, 0xad, 0xbe, 0xef}
	value := big.NewInt(123456)

	params := MethodParameters{
		Calldata: calldata,
		Value:    value,
	}

	if len(params.Calldata) != 4 {
		t.Errorf("Expected Calldata length 4, got %d", len(params.Calldata))
	}

	if params.Value.Cmp(value) != 0 {
		t.Errorf("Expected Value %v, got %v", value, params.Value)
	}
}

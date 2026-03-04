package sdkv4

import (
	"math/big"
)

// MethodParameters represents the input parameters for a contract method call.
// It typically includes the encoded calldata and the value of Ether (or token) to send.
type MethodParameters struct {
	// Calldata is the ABI-encoded bytes representing the function and arguments.
	Calldata []byte

	// Value is the amount of Ether (in wei) to send along with the method call.
	Value *big.Int
}

// ToHex converts a big.Int value to its hexadecimal string representation.
// The returned string is prefixed with "0x" and is zero-padded to have an even length.
// If the input is nil, it returns "0x00".
//
// Example usage:
//
//	i := big.NewInt(255)
//	hexStr := ToHex(i) // "0xFF"
func ToHex(i *big.Int) string {
	if i == nil {
		return "0x00"
	}

	hex := i.Text(16)
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	return "0x" + hex
}

package utils

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestCalculatePositionKey(t *testing.T) {
	owner := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	tickLower := int(-120)
	tickUpper := int(150)
	salt := [32]byte{}

	key := ComputePositionKey(owner, tickLower, tickUpper, salt)

	expectedHex := "d9c6351d73ec1a9706a2e596213ff513744be32f0a6f58fa0b423f43619f95a3"
	if expectedHex != "" {
		expectedBytes, err := hex.DecodeString(expectedHex)
		if err != nil {
			t.Fatalf("decode expected hex: %v", err)
		}
		if !bytes.Equal(key[:], expectedBytes) {
			t.Errorf("position key mismatch, got %x, want %x", key, expectedBytes)
		}
	}
}

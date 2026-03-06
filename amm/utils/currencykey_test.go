package utils

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestComputeCurrencySlotKey(t *testing.T) {
	owner := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	currency := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	key1 := ComputeCurrencySlotKey(owner, currency)
	key2 := ComputeCurrencySlotKey(owner, currency)

	// Deterministic: calling twice with the same inputs produces the same key
	assert.Equal(t, key1, key2, "Keys should be deterministic and equal")

	// Changing currency should produce a different key
	otherCurrency := common.HexToAddress("0x1111111111111111111111111111111111111111")
	key3 := ComputeCurrencySlotKey(owner, otherCurrency)
	assert.NotEqual(t, key1, key3, "Keys for different currency should differ")
}

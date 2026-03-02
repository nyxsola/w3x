package libraries_test

import (
	"math/big"
	"testing"

	"github.com/aicora/go-uniswap/core/libraries"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TestCurrencyDeltaBasic tests the basic ApplyDelta and GetDelta functionality.
func TestCurrencyDeltaBasic(t *testing.T) {
	cd := libraries.NewCurrencyDelta()

	owner := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	currency := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	// Apply a delta of 100
	prev, next := cd.ApplyDelta(owner, currency, big.NewInt(100))

	// Initial previous delta should be zero
	assert.Equal(t, 0, prev.Cmp(big.NewInt(0)), "previous delta should be zero")
	assert.Equal(t, 0, next.Cmp(big.NewInt(100)), "next delta should equal the applied delta")

	// GetDelta should return the same value
	current := cd.GetDelta(owner, currency)
	assert.Equal(t, 0, current.Cmp(next), "GetDelta should match last applied delta")
}

// TestCurrencyDeltaNegativeDelta tests applying negative deltas.
func TestCurrencyDeltaNegativeDelta(t *testing.T) {
	cd := libraries.NewCurrencyDelta()

	owner := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	currency := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	// Apply a positive delta
	_, next := cd.ApplyDelta(owner, currency, big.NewInt(100))
	assert.Equal(t, 0, next.Cmp(big.NewInt(100)), "next delta should be 100")

	// Apply a negative delta
	prev2, next2 := cd.ApplyDelta(owner, currency, big.NewInt(-40))
	assert.Equal(t, 0, prev2.Cmp(next), "previous delta should equal last next")
	assert.Equal(t, 0, next2.Cmp(big.NewInt(60)), "next delta should be cumulative (100-40=60)")
}

// TestCurrencyDeltaMultipleOwners ensures deltas are isolated per owner.
func TestCurrencyDeltaMultipleOwners(t *testing.T) {
	cd := libraries.NewCurrencyDelta()

	owner1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	owner2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	currency := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	// Apply deltas to different owners
	cd.ApplyDelta(owner1, currency, big.NewInt(100))
	cd.ApplyDelta(owner2, currency, big.NewInt(200))

	assert.Equal(t, 0, cd.GetDelta(owner1, currency).Cmp(big.NewInt(100)), "owner1 delta should be 100")
	assert.Equal(t, 0, cd.GetDelta(owner2, currency).Cmp(big.NewInt(200)), "owner2 delta should be 200")
}

// TestCurrencyDeltaDifferentCurrencies ensures the same owner with different currencies are isolated.
func TestCurrencyDeltaDifferentCurrencies(t *testing.T) {
	cd := libraries.NewCurrencyDelta()

	owner := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	currency1 := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	currency2 := common.HexToAddress("0x1111111111111111111111111111111111111111")

	cd.ApplyDelta(owner, currency1, big.NewInt(50))
	cd.ApplyDelta(owner, currency2, big.NewInt(75))

	assert.Equal(t, 0, cd.GetDelta(owner, currency1).Cmp(big.NewInt(50)), "currency1 delta should be 50")
	assert.Equal(t, 0, cd.GetDelta(owner, currency2).Cmp(big.NewInt(75)), "currency2 delta should be 75")
}
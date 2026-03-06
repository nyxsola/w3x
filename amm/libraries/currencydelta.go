package libraries

import (
	"math/big"

	"github.com/aicora/w3x/amm/utils"
	"github.com/ethereum/go-ethereum/common"
)

// CurrencyDelta provides a centralised equivalent of Solidity's CurrencyDelta library.
//
// It maintains a mapping of account and currency pairs to their transient balance deltas.
// This is a deterministic, in-memory representation of what Solidity achieves with
// assembly `tload`/`tstore` and `_computeSlot`.
type CurrencyDelta struct {
	// deltas maps the keccak256(owner + currency) key to the delta amount.
	// The key is generated using utils.ComputeCurrencySlotKey.
	deltas map[[32]byte]*big.Int
}

// NewCurrencyDelta creates a new CurrencyDelta instance with an initialized map.
func NewCurrencyDelta() *CurrencyDelta {
	return &CurrencyDelta{
		deltas: make(map[[32]byte]*big.Int),
	}
}

// GetDelta returns the current balance delta for a given account and currency.
//
// Parameters:
//   - owner:   the Ethereum address of the account whose delta is being queried
//   - currency: the Ethereum address representing the currency
//
// Returns:
//   - *big.Int: the current delta; returns 0 if no delta has been recorded yet
//
// Example usage:
//
//	delta := cd.GetDelta(userAddress, tokenAddress)
func (cd *CurrencyDelta) GetDelta(owner, currency common.Address) *big.Int {
	slot := utils.ComputeCurrencySlotKey(owner, currency)
	if val, ok := cd.deltas[slot]; ok {
		return new(big.Int).Set(val)
	}
	return big.NewInt(0)
}

// ApplyDelta applies a delta adjustment to a given account and currency pair.
//
// This mimics Solidity's applyDelta function by updating the internal map and
// returning both the previous and the next delta values.
//
// Parameters:
//   - owner:   the Ethereum address of the account to update
//   - currency: the Ethereum address representing the currency
//   - delta:    the delta to apply (can be positive or negative)
//
// Returns:
//   - previous: the prior delta before applying the adjustment
//   - next:     the updated delta after applying the adjustment
//
// Example usage:
//
//	prev, next := cd.ApplyDelta(userAddress, tokenAddress, big.NewInt(100))
func (cd *CurrencyDelta) ApplyDelta(owner, currency common.Address, delta *big.Int) (previous, next *big.Int) {
	slot := utils.ComputeCurrencySlotKey(owner, currency)
	previous = cd.GetDelta(owner, currency)
	next = new(big.Int).Add(previous, delta)
	cd.deltas[slot] = new(big.Int).Set(next)
	return previous, next
}

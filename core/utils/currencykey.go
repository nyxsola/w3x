package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeCurrencySlotKey computes a unique 32-byte key for tracking a user's currency delta.
//
// This function is a Go equivalent of Solidity's _computeSlot in the CurrencyDelta library.
// It concatenates the 20-byte owner address and the 20-byte currency address, then hashes
// the 40-byte packed input using Keccak256 to produce a unique storage slot key.
//
// Parameters:
//   - owner:   the Ethereum address of the account whose delta is being tracked (20 bytes)
//   - currency: the Ethereum address representing the currency or token (20 bytes)
//
// Returns:
//   - [32]byte: a deterministic 32-byte key that can be used as a map key or storage slot
//
// Example usage:
//
//    key := ComputeCurrencySlotKey(userAddress, tokenAddress)
//    deltaMap[key] = newValue
//
func ComputeCurrencySlotKey(owner, currency common.Address) [32]byte {
	buf := make([]byte, 0, 40)

	// Append owner address (20 bytes)
	buf = append(buf, owner.Bytes()...)

	// Append currency address (20 bytes)
	buf = append(buf, currency.Bytes()...)

	// Compute Keccak256 hash of the packed owner + currency bytes
	return crypto.Keccak256Hash(buf)
}
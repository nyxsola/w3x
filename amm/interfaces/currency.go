package interfaces

import "github.com/ethereum/go-ethereum/common"

// ICurrency defines the interface for a blockchain currency/token.
type ICurrency interface {
	// IsNative returns true if this currency is the native token of the chain.
	IsNative() bool

	// Address returns the canonical lowercase address of the token.
	Address() common.Address

	// ChainId returns the chain ID of the currency.
	ChainId() uint

	// Decimals returns the number of decimals the token uses.
	Decimals() uint8

	// Symbol returns the token symbol (e.g., "ETH", "USDT").
	Symbol() string

	// Name returns the token name (e.g., "Ethereum", "Tether").
	Name() string

	// Equal checks whether two currencies are identical (same chain and address).
	Equal(other ICurrency) bool

	// Lt compares two currencies by address lexicographically.
	// Returns an error if currencies are on different chains or have the same address.
	Lt(other ICurrency) (bool, error)
}

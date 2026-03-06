package libraries

import (
	"github.com/aicora/w3x/amm/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidAddressFormat is returned when a currency address is invalid.
	ErrInvalidAddressFormat = errors.New("invalid address format")

	// ErrSameAddress is returned when two currencies have the same address but ordering is required.
	ErrSameAddress = errors.New("same address")
)

var (
	// ZeroAddress is the canonical zero address in Ethereum.
	ZeroAddress = common.Address{}
)

// isNative checks if the given address represents a native token (ETH, BNB, etc.).
//
// Parameters:
//   - address: the token address to check
//
// Returns:
//   - true if the address is empty, 0x0, or the canonical zero address.
func isNative(address common.Address) bool {
	return address == ZeroAddress
}

// Currency implements ICurrency and represents a blockchain token or native currency.
type Currency struct {
	isNative bool           // true if the currency is the native token of the chain
	address  common.Address // canonical token address
	chainId  uint           // chain ID
	decimals uint8          // number of token decimals
	symbol   string         // token symbol
	name     string         // token name
}

// NewCurrency creates a new Currency instance.
//
// Parameters:
//   - chainID: the blockchain chain ID
//   - address: token address; empty or zero address indicates native token
//   - decimals: number of token decimals
//   - symbol: token symbol
//   - name: token name
//
// Returns:
//   - ICurrency instance representing the token
func NewCurrency(chainID uint, address common.Address, decimals uint8, symbol string, name string) *Currency {
	return &Currency{
		isNative: isNative(address),
		address:  address,
		chainId:  chainID,
		decimals: decimals,
		symbol:   symbol,
		name:     name,
	}
}

// IsNative returns true if this currency is the native token of the chain.
func (c *Currency) IsNative() bool {
	return c.isNative
}

// Address returns the canonical lowercase address of the token.
func (c *Currency) Address() common.Address {
	return c.address
}

// ChainId returns the chain ID of the currency.
func (c *Currency) ChainId() uint {
	return c.chainId
}

// Decimals returns the number of decimals used by the token.
func (c *Currency) Decimals() uint8 {
	return c.decimals
}

// Symbol returns the token symbol (e.g., "ETH").
func (c *Currency) Symbol() string {
	return c.symbol
}

// Name returns the token name (e.g., "Ethereum").
func (c *Currency) Name() string {
	return c.name
}

// Equal checks whether two currencies are identical (same chain ID and address).
func (c *Currency) Equal(other interfaces.ICurrency) bool {
	return c.chainId == other.ChainId() && c.address == other.Address()
}

// Lt compares two currencies lexicographically by address.
//
// Returns:
//   - true if c < other
//   - ErrSameAddress if currencies have the same address
func (c *Currency) Lt(other interfaces.ICurrency) (bool, error) {
	if c.address == other.Address() {
		return false, ErrSameAddress
	}
	return c.address.Big().Cmp(other.Address().Big()) < 0, nil
}

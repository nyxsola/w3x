package sdkcore

import (
	"github.com/ethereum/go-ethereum/common"
)

// CheckValidAddress validates the basic format of an Ethereum address.
//
// It checks whether the address:
//   - is a valid hexadecimal string
//   - starts with the "0x" prefix
//   - has a total length of 42 characters (including "0x")
//
// Note: This function does not enforce EIP-55 checksum.
//
// Parameters:
//   - address: the Ethereum address string to validate
//
// Returns:
//   - the original address if valid
//   - an error if the address format is invalid
func CheckValidAddress(address string) (string, error) {
	if !common.IsHexAddress(address) {
		return "", ErrInvalidAddressFormat
	}
	return address, nil
}

// ValidateAndParseAddress validates and returns the checksummed version of an Ethereum address.
//
// This function performs the following steps:
//  1. Checks whether the input address is a valid hex address
//  2. Converts the address to the canonical EIP-55 checksum format using go-ethereum's HexToAddress
//
// Parameters:
//   - address: the Ethereum address string to validate and parse
//
// Returns:
//   - the checksummed Ethereum address string (EIP-55 format) if valid
//   - an error if the address is invalid
func ValidateAndParseAddress(address string) (string, error) {
	if !common.IsHexAddress(address) {
		return "", ErrInvalidAddressFormat
	}
	return common.HexToAddress(address).Hex(), nil
}

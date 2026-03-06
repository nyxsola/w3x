package libraries

import (
	"encoding/hex"
	"math/big"

	"github.com/aicora/go-uniswap/amm/interfaces"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// PoolId represents the unique identifier of a pool.
//
// It corresponds to Solidity's `bytes32` PoolId, which is computed
// as `keccak256(abi.encode(poolKey))` on-chain. This ID uniquely
// identifies a pool based on its key parameters.
type PoolId [32]byte

// String returns a hex-encoded string representation of the PoolId.
func (p PoolId) String() string {
	return hex.EncodeToString(p[:])
}

// PoolKey defines the key parameters that uniquely identify a pool.
//
// The ordering of Currency0 and Currency1 is normalized (Currency0 < Currency1 by address)
// to ensure the PoolId is deterministic and consistent with Uniswap contracts.
//
// Fields:
//   - Currency0: the lower currency by address
//   - Currency1: the higher currency by address
//   - Fee: the pool's LP fee, must fit uint24
//   - TickSpacing: tick spacing, must fit int24
//   - Hooks: optional hooks contract that implements pool callbacks
type PoolKey struct {
	Currency0   interfaces.ICurrency
	Currency1   interfaces.ICurrency
	Fee         LPFee
	TickSpacing int
	Hooks       interfaces.IHooks
}

// ToId computes the PoolId for this PoolKey.
//
// The returned PoolId is equivalent to Solidity's `keccak256(abi.encode(poolKey))`.
//
// Behavior:
//   - Ensures Currency0 and Currency1 are sorted by address
//   - Packs the struct fields using Ethereum ABI encoding
//   - Computes keccak256 hash over the ABI-encoded bytes
//
// Returns:
//   - PoolId: 32-byte identifier of the pool
//   - error: if ABI packing fails
func (p *PoolKey) ToId() (PoolId, error) {
	// Sort currencies to ensure deterministic order (token0 < token1)
	if p.Currency1.Address().Big().Cmp(p.Currency0.Address().Big()) < 0 {
		p.Currency0, p.Currency1 = p.Currency1, p.Currency0
	}

	// Define ABI types matching Solidity PoolKey struct
	currency0, err := abi.NewType("address", "", nil)
	if err != nil {
		return PoolId{}, err
	}
	currency1, err := abi.NewType("address", "", nil)
	if err != nil {
		return PoolId{}, err
	}
	fee, err := abi.NewType("uint24", "", nil)
	if err != nil {
		return PoolId{}, err
	}
	tickSpacing, err := abi.NewType("int24", "", nil)
	if err != nil {
		return PoolId{}, err
	}
	hooks, err := abi.NewType("address", "", nil)
	if err != nil {
		return PoolId{}, err
	}

	// Pack the fields using ABI encoding in Solidity struct order
	arguments := abi.Arguments{
		{Type: currency0},
		{Type: currency1},
		{Type: fee},
		{Type: tickSpacing},
		{Type: hooks},
	}

	hooksAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
	if p.Hooks != nil {
		hooksAddr = p.Hooks.Address()
	}

	packed, err := arguments.Pack(
		p.Currency0.Address(),
		p.Currency1.Address(),
		big.NewInt(int64(p.Fee)),
		big.NewInt(int64(p.TickSpacing)),
		hooksAddr,
	)
	if err != nil {
		return PoolId{}, err
	}

	// Compute keccak256 hash to produce PoolId
	return PoolId(crypto.Keccak256Hash(packed)), nil
}

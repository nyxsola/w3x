package utils

import (
	"math/big"
)

var Q256 = new(big.Int).Lsh(big.NewInt(1), 256)
var Q128 = new(big.Int).Lsh(big.NewInt(1), 128)
var Q96 = new(big.Int).Lsh(big.NewInt(1), 96) // 2^96
var Q32 = big.NewInt(1 << 32)
var MaxUint160 = new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(160), nil), big.NewInt(1))
var MaxUint256, _ = new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

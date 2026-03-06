package interfaces

import "github.com/ethereum/go-ethereum/common"

type IHooks interface {
	Address() common.Address
}

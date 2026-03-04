package sdkcore

import "github.com/ethereum/go-ethereum/common"

// Known WETH9 implementation addresses, used in our implementation of Ether#wrapped
var WETH9 = map[uint]*Currency{
	1:        NewCurrency(1, common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), 18, "WETH", "Wrapped Ether"),
	11155111: NewCurrency(11155111, common.HexToAddress("0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14"), 18, "WETH", "Wrapped Ether"),
	3:        NewCurrency(3, common.HexToAddress("0xc778417E063141139Fce010982780140Aa0cD5Ab"), 18, "WETH", "Wrapped Ether"),
	4:        NewCurrency(4, common.HexToAddress("0xc778417E063141139Fce010982780140Aa0cD5Ab"), 18, "WETH", "Wrapped Ether"),
	5:        NewCurrency(5, common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"), 18, "WETH", "Wrapped Ether"),
	42:       NewCurrency(42, common.HexToAddress("0xd0A1E359811322d97991E03f863a0C30C2cF029C"), 18, "WETH", "Wrapped Ether"),

	10: NewCurrency(10, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	69: NewCurrency(69, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),

	42161:  NewCurrency(42161, common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1"), 18, "WETH", "Wrapped Ether"),
	421611: NewCurrency(421611, common.HexToAddress("0xB47e6A5f8b33b3F17603C83a0535A9dcD7E32681"), 18, "WETH", "Wrapped Ether"),
	421614: NewCurrency(421614, common.HexToAddress("0x980B62Da83eFf3D4576C647993b0c1D7faf17c73"), 18, "WETH", "Wrapped Ether"),

	8453:  NewCurrency(8453, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	84532: NewCurrency(84532, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),

	7777777: NewCurrency(7777777, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	81457:   NewCurrency(81457, common.HexToAddress("0x4300000000000000000000000000000000000004"), 18, "WETH", "Wrapped Ether"),

	324:  NewCurrency(324, common.HexToAddress("0x5AEa5775959fBC2557Cc8789bC1bf90A239D9a91"), 18, "WETH", "Wrapped Ether"),
	480:  NewCurrency(480, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	1301: NewCurrency(1301, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	130:  NewCurrency(130, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),

	1868:  NewCurrency(1868, common.HexToAddress("0x4200000000000000000000000000000000000006"), 18, "WETH", "Wrapped Ether"),
	59144: NewCurrency(59144, common.HexToAddress("0xe5D7C2a44FfDDf6b295A15c148167daaAf5Cf34f"), 18, "WETH", "Wrapped Ether"),
}

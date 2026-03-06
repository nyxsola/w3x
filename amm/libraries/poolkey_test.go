package libraries

import (
	"testing"

	"github.com/aicora/go-uniswap/amm/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type mockHooks struct {
	addr common.Address
}

func (m mockHooks) Address() common.Address {
	return m.addr
}

type mockCurrency struct {
	chainId  uint
	address  common.Address
	decimals uint
}

func TestPoolKey_ToId(t *testing.T) {
	tokenA := NewCurrency(1, common.HexToAddress("0x558AFaF6FeF52395D558F9fc1ab18A08C7A7548b"), 18, "A", "TokenA")
	tokenB := NewCurrency(1, common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), 18, "B", "TokenB")
	hooks := mockHooks{addr: common.HexToAddress("0x0000000000000000000000000000000000000000")}

	tests := []struct {
		name      string
		currency0 interfaces.ICurrency
		currency1 interfaces.ICurrency
		fee       LPFee
		tickSpace int
		hooks     interfaces.IHooks
		wantSwap  bool
	}{
		{
			name:      "normal order",
			currency0: tokenA,
			currency1: tokenB,
			fee:       10000,
			tickSpace: 60,
			hooks:     hooks,
			wantSwap:  false,
		},
		{
			name:      "reverse order",
			currency0: tokenB,
			currency1: tokenA,
			fee:       500,
			tickSpace: 10,
			hooks:     hooks,
			wantSwap:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk := &PoolKey{
				Currency0:   tt.currency0,
				Currency1:   tt.currency1,
				Fee:         tt.fee,
				TickSpacing: tt.tickSpace,
				Hooks:       tt.hooks,
			}

			pid, err := pk.ToId()
			require.NoError(t, err, "ToId should not return error")
			if tt.wantSwap {
				require.True(t, pk.Currency0.Address().Big().Cmp(pk.Currency1.Address().Big()) < 0,
					"Currency0 should be smaller than Currency1 after ToId")
			} else {
				require.True(t, pk.Currency0.Address().Big().Cmp(pk.Currency1.Address().Big()) < 0,
					"Currency0 should remain smaller than Currency1")
			}

			pid2, err := pk.ToId()
			require.NoError(t, err)
			require.Equal(t, pid, pid2, "PoolId should be deterministic and repeatable")
		})
	}
}

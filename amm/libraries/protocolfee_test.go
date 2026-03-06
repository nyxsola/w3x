package libraries

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtocolFeeCreationAndExtraction(t *testing.T) {
	pf, err := NewProtocolFee(500, 800)
	require.NoError(t, err)
	require.Equal(t, uint16(500), pf.ZeroForOne())
	require.Equal(t, uint16(800), pf.OneForZero())
	require.True(t, pf.IsValid())

	_, err = NewProtocolFee(1001, 500)
	require.ErrorIs(t, err, ErrInvalidProtocolFee)
	_, err = NewProtocolFee(500, 2000)
	require.ErrorIs(t, err, ErrInvalidProtocolFee)
}

func TestProtocolFeeIsValid(t *testing.T) {
	pf, _ := NewProtocolFee(0, 0)
	require.True(t, pf.IsValid())

	pf, _ = NewProtocolFee(MaxProtocolFee, MaxProtocolFee)
	require.True(t, pf.IsValid())

	invalidPf := ProtocolFee(1002 | (1002 << 12))
	require.False(t, invalidPf.IsValid())
}

func TestCalculateSwapFee(t *testing.T) {
	lp, _ := NewFee(100_000) // 10%
	// protocolFee = 500 pips = 0.05%
	swapFee := CalculateSwapFee(500, lp)
	// 0.0005 + 0.1 - 0.0005*0.1 ≈ 0.10045 -> 100_450 pips
	expected := 500 + 100_000 - (500*100_000)/PipsDenominator
	require.Equal(t, expected, swapFee)

	// protocolFee = 0, lpFee = 0
	lp0, _ := NewFee(0)
	require.Equal(t, uint32(0), CalculateSwapFee(0, lp0))

	// protocolFee = MaxProtocolFee, lpFee = MaxLPFee
	lpMax, _ := NewFee(MaxLPFee)
	expectedMax := uint32(MaxProtocolFee) + MaxLPFee - (uint32(MaxProtocolFee)*MaxLPFee)/PipsDenominator
	require.Equal(t, expectedMax, CalculateSwapFee(MaxProtocolFee, lpMax))
}

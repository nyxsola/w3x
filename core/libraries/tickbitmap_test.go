package libraries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTickBitmap_FlipAndIsInitialized(t *testing.T) {
	tb := NewTickBitmap()
	tick := int(10)
	tickSpacing := int(1)

	if tb.IsInitialized(tick, tickSpacing) {
		t.Fatal("tick should not be initialized initially")
	}

	if err := tb.FlipTick(tick, tickSpacing); err != nil {
		t.Fatalf("FlipTick error: %v", err)
	}

	if !tb.IsInitialized(tick, tickSpacing) {
		t.Fatal("tick should be initialized after FlipTick")
	}

	if err := tb.FlipTick(tick, tickSpacing); err != nil {
		t.Fatalf("FlipTick error: %v", err)
	}
	if tb.IsInitialized(tick, tickSpacing) {
		t.Fatal("tick should be uninitialized after second FlipTick")
	}
}

func TestTickBitmap_FlipTick(t *testing.T) {
	tb := NewTickBitmap()
	tickSpacing := int(10)
	tick := int(20)

	// Flip tick to initialized
	if err := tb.FlipTick(tick, tickSpacing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Flip tick again to uninitialized
	if err := tb.FlipTick(tick, tickSpacing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check uninitialized
	compressed := tick / tickSpacing
	wordPos := int16(compressed >> 8)
	if word, ok := tb.words[wordPos]; ok {
		if word.Sign() != 0 {
			t.Errorf("expected word to be zero after flipping back, got %v", word)
		}
	}
}

func TestNextInitializedTickWithinOneWord(t *testing.T) {
	tb := NewTickBitmap()

	tickSpacing := 1

	tb.FlipTick(5, tickSpacing)
	tb.FlipTick(10, tickSpacing)
	tb.FlipTick(250, tickSpacing)

	tests := []struct {
		name        string
		tick        int
		lte         bool
		wantNext    int
		wantInit    bool
	}{
		{
			name:     "lte=true, tick below first initialized",
			tick:     2,
			lte:      true,
			wantNext: 0,
			wantInit: false,
		},
		{
			name:     "lte=true, tick exactly on initialized",
			tick:     10,
			lte:      true,
			wantNext: 10,
			wantInit: true,
		},
		{
			name:     "lte=true, tick between initialized",
			tick:     7,
			lte:      true,
			wantNext: 5,
			wantInit: true,
		},
		{
			name:     "lte=false, tick below first initialized",
			tick:     2,
			lte:      false,
			wantNext: 5,
			wantInit: true,
		},
		{
			name:     "lte=false, tick exactly on initialized",
			tick:     10,
			lte:      false,
			wantNext: 250,
			wantInit: true,
		},
		{
			name:     "lte=false, tick between initialized",
			tick:     6,
			lte:      false,
			wantNext: 10,
			wantInit: true,
		},
		{
			name:     "lte=true, tick above highest initialized",
			tick:     255,
			lte:      true,
			wantNext: 250,
			wantInit: true,
		},
		{
			name:     "lte=false, tick above highest initialized",
			tick:     255,
			lte:      false,
			wantNext: 511, 
			wantInit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, init, err := tb.NextInitializedTickWithinOneWord(tt.tick, tickSpacing, tt.lte)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantNext, got)
			assert.Equal(t, tt.wantInit, init)
		})
	}
}
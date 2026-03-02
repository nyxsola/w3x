package libraries

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFee(t *testing.T) {
	fee, err := NewFee(3000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fee.Value() != 3000 {
		t.Fatalf("expected value 3000, got %d", fee.Value())
	}
}

func TestNewFee_TooLarge(t *testing.T) {
	_, err := NewFee(MaxLPFee + 1)
	if !errors.Is(err, ErrFeeTooLarge) {
		t.Fatalf("expected ErrFeeTooLarge, got %v", err)
	}
}

func TestNewDynamicFee(t *testing.T) {
	fee := NewDynamicFee()

	if !fee.IsDynamic() {
		t.Fatal("expected dynamic fee")
	}

	if fee.Value() != 0 {
		t.Fatalf("expected value 0, got %d", fee.Value())
	}
}

func TestOverrideFlag(t *testing.T) {
	fee, _ := NewFee(5000)

	fee = fee.WithOverride()

	if !fee.IsOverride() {
		t.Fatal("expected override flag set")
	}

	cleared := fee.RemoveOverride()

	if cleared.IsOverride() {
		t.Fatal("expected override flag cleared")
	}

	if cleared.Value() != 5000 {
		t.Fatalf("expected value 5000, got %d", cleared.Value())
	}
}

func TestValidate_StaticFee(t *testing.T) {
	fee, _ := NewFee(1000)

	if err := fee.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_DynamicFee(t *testing.T) {
	fee := NewDynamicFee()

	if err := fee.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_TooLargeValue(t *testing.T) {
	fee := LPFee(MaxLPFee + 1)

	err := fee.Validate()
	if !errors.Is(err, ErrFeeTooLarge) {
		t.Fatalf("expected ErrFeeTooLarge, got %v", err)
	}
}

func TestInitialValue_Static(t *testing.T) {
	fee, _ := NewFee(2500)

	val, err := fee.InitialValue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != 2500 {
		t.Fatalf("expected 2500, got %d", val)
	}
}

func TestInitialValue_Dynamic(t *testing.T) {
	fee := NewDynamicFee()

	val, err := fee.InitialValue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != 0 {
		t.Fatalf("expected 0, got %d", val)
	}
}

func TestLPFee_RemoveOverrideFlagAndValidate(t *testing.T) {

	tests := []struct {
		name        string
		input       LPFee
		expectFee   LPFee
		expectError bool
	}{
		{
			name:        "valid fee without override",
			input:       LPFee(3000),
			expectFee:   LPFee(3000),
			expectError: false,
		},
		{
			name:        "valid fee with override flag",
			input:       LPFee(overrideFlag | 3000),
			expectFee:   LPFee(3000),
			expectError: false,
		},
		{
			name:        "zero fee",
			input:       LPFee(0),
			expectFee:   LPFee(0),
			expectError: false,
		},
		{
			name:        "only override flag",
			input:       LPFee(overrideFlag),
			expectFee:   LPFee(0),
			expectError: false,
		},
		{
			name:        "invalid fee over max",
			input:       LPFee(overrideFlag | (MaxLPFee + 1)),
			expectFee:   0,
			expectError: true,
		},
		{
			name:        "invalid fee without override",
			input:       LPFee(MaxLPFee + 1),
			expectFee:   0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fee, err := tt.input.RemoveOverrideFlagAndValidate()

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectFee, fee)
		})
	}
}
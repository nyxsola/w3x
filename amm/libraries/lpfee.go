package libraries

import "github.com/pkg/errors"

// LPFee is a uint24-encoded fee structure.
//
// Layout (24 bits total):
//
//   bit 23 (0x800000) -> dynamic fee flag
//   bit 22 (0x400000) -> override flag
//   bits 0-21         -> actual fee value
//
// Fee unit:
//   1 = 1e-6 (one millionth)
//   MaxLPFee = 1_000_000 (100%)
//
// This encoding mirrors Uniswap v4's fee model.
//

const (
	// maxUint24 represents the maximum value of a uint24.
	maxUint24 uint32 = 0xFFFFFF

	// dynamicFlag marks a pool as dynamic-fee enabled.
	dynamicFlag uint32 = 1 << 23 // 0x800000

	// overrideFlag marks a swap as using an override fee.
	overrideFlag uint32 = 1 << 22 // 0x400000

	// valueMask extracts the actual fee value (lower 22 bits).
	valueMask uint32 = overrideFlag - 1

	// MaxLPFee defines the maximum allowed LP fee (1e6 = 100%).
	MaxLPFee uint32 = 1_000_000
)

var (
	// ErrFeeTooLarge is returned when the fee exceeds MaxLPFee.
	ErrFeeTooLarge = errors.New("lp fee exceeds maximum allowed value")
)

// LPFee represents a uint24-encoded liquidity provider fee.
type LPFee uint32

// NewFee creates a static LP fee.
//
// Returns ErrFeeTooLarge if value > MaxLPFee.
func NewFee(value uint32) (LPFee, error) {
	if value > MaxLPFee {
		return 0, ErrFeeTooLarge
	}
	return LPFee(value), nil
}

// NewDynamicFee creates a dynamic-fee LP pool.
func NewDynamicFee() LPFee {
	return LPFee(dynamicFlag)
}

// Raw returns the raw uint24 representation.
func (f LPFee) Raw() uint32 {
	return uint32(f) & maxUint24
}

// IsDynamic reports whether the dynamic fee flag is set.
func (f LPFee) IsDynamic() bool {
	return uint32(f)&dynamicFlag != 0
}

// IsOverride reports whether the override flag is set.
func (f LPFee) IsOverride() bool {
	return uint32(f)&overrideFlag != 0
}

// Value extracts the actual fee value (lower 22 bits).
func (f LPFee) Value() uint32 {
	return uint32(f) & valueMask
}

// RemoveOverride clears the override flag.
func (f LPFee) RemoveOverride() LPFee {
	return LPFee(uint32(f) &^ overrideFlag)
}

func (f LPFee) RemoveOverrideFlagAndValidate() (LPFee, error) {
	fee := f.RemoveOverride()
	if err := fee.Validate(); err != nil {
		return 0, err
	}
	return fee, nil
}

// WithOverride sets the override flag.
func (f LPFee) WithOverride() LPFee {
	return LPFee(uint32(f) | overrideFlag)
}

// Validate verifies that the LP fee is within allowed bounds.
//
// Rules:
//   - Must fit within uint24
//   - If static fee, Value() must be <= MaxLPFee
//
// Dynamic fee pools bypass static fee range checks.
func (f LPFee) Validate() error {
	raw := f.Raw()

	if raw > maxUint24 {
		return ErrFeeTooLarge
	}

	if !f.IsDynamic() && f.Value() > MaxLPFee {
		return ErrFeeTooLarge
	}

	return nil
}

// InitialValue returns the initial LP fee for pool initialization.
//
// For dynamic-fee pools, returns 0.
// For static-fee pools, validates and returns Value().
func (f LPFee) InitialValue() (uint32, error) {
	if f.IsDynamic() {
		return 0, nil
	}

	if err := f.Validate(); err != nil {
		return 0, err
	}

	return f.Value(), nil
}

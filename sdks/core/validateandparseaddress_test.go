package sdkcore

import (
	"testing"
)

func TestCheckValidAddress(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantValid bool
	}{
		{
			name:      "valid checksummed",
			address:   "0x52908400098527886E0F7030069857D2E4169EE7",
			wantValid: true,
		},
		{
			name:      "valid lowercase",
			address:   "0x52908400098527886e0f7030069857d2e4169ee7",
			wantValid: true,
		},
		{
			name:      "valid missing 0x",
			address:   "52908400098527886E0F7030069857D2E4169EE7",
			wantValid: true,
		},
		{
			name:      "invalid length",
			address:   "0x123",
			wantValid: false,
		},
		{
			name:      "invalid hex",
			address:   "0xZZZZZZ00098527886E0F7030069857D2E4169EE7",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CheckValidAddress(tt.address)
			if (err == nil) != tt.wantValid {
				t.Errorf("CheckValidAddress(%s) valid=%v, got error: %v", tt.address, tt.wantValid, err)
			}
		})
	}
}

func TestValidateAndParseAddress(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOutput  string
		expectError bool
	}{
		{
			name:        "valid checksummed",
			input:       "0x52908400098527886E0F7030069857D2E4169EE7",
			wantOutput:  "0x52908400098527886E0F7030069857D2E4169EE7",
			expectError: false,
		},
		{
			name:        "valid lowercase",
			input:       "0x52908400098527886e0f7030069857d2e4169ee7",
			wantOutput:  "0x52908400098527886E0F7030069857D2E4169EE7",
			expectError: false,
		},
		{
			name:        "valid uppercase",
			input:       "0X52908400098527886E0F7030069857D2E4169EE7",
			wantOutput:  "0x52908400098527886E0F7030069857D2E4169EE7",
			expectError: false,
		},
		{
			name:        "invalid address",
			input:       "0x123",
			wantOutput:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndParseAddress(tt.input)
			if (err != nil) != tt.expectError {
				t.Fatalf("ValidateAndParseAddress(%s) error = %v, expectError=%v", tt.input, err, tt.expectError)
			}
			if !tt.expectError && got != tt.wantOutput {
				t.Errorf("ValidateAndParseAddress(%s) = %s, want %s", tt.input, got, tt.wantOutput)
			}
		})
	}
}

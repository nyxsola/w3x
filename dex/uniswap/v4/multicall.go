package uniswapsdkv4

import (
	"strings"

	uniswapv3periphery "github.com/aicora/go-uniswap/dex/uniswap/v3-periphery"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/tidwall/gjson"
)

var (
	// ErrUnexpectedMulticallOutput indicates that the decoded multicall output
	// does not have the expected number of return values.
	ErrUnexpectedMulticallOutput = errors.New("unexpected multicall output")

	// ErrUnexpectedMulticallOutputType indicates that the decoded multicall output
	// is not of the expected type [][]byte.
	ErrUnexpectedMulticallOutputType = errors.New("unexpected type for multicall output")

	// ErrShortCalldata indicates that the encoded calldata is too short.
	ErrShortCalldata = errors.New("encoded calldata too short")
)

// MulticallABI is the parsed ABI for the IMulticall interface.
// It is used for encoding and decoding multicall calls.
var MulticallABI abi.ABI

func init() {
	// Extract the "abi" field from the embedded IMulticall JSON using gjson.
	// gjson allows high-performance retrieval of nested JSON fields without unmarshaling the entire JSON.
	abiJSON := gjson.GetBytes(uniswapv3periphery.IMulticall, "abi")

	// Parse the ABI JSON into go-ethereum's abi.ABI type.
	// Panic if the ABI is invalid, as this should never fail with a correct JSON.
	var err error
	MulticallABI, err = abi.JSON(strings.NewReader(abiJSON.Raw))
	if err != nil {
		panic(err)
	}
}

// EncodeMulticall encodes one or multiple calldata entries into a single
// multicall call data payload.
//
// If calldataList contains only one element, it returns it directly
// without encoding, otherwise it packs the list into the "multicall" method call.
func EncodeMulticall(calldataList [][]byte) ([]byte, error) {
	return MulticallABI.Pack("multicall", calldataList)
}

// DecodeMulticall decodes the input calldata of a `multicall(bytes[])` function call.
//
// The provided calldata must be the full ABI-encoded payload, including the
// 4-byte function selector prefix. This function will automatically strip
// the selector before decoding the arguments.
//
// According to the ABI specification, the `multicall` method takes a single
// parameter of type `bytes[]`, which represents an array of encoded contract
// method calls. The decoded result is returned as a slice of byte slices
// ([][]byte), where each element corresponds to one sub-call payload.
//
// Example calldata layout:
//
//	0x<4-byte selector>
//	  <ABI-encoded bytes[]>
//
// Returns:
//   - [][]byte: The decoded array of sub-call calldata.
//   - error:    An error if the calldata is malformed, too short,
//     ABI decoding fails, or the output type is unexpected.
//
// Errors:
//   - ErrShortCalldata                if calldata is shorter than 4 bytes
//   - ErrUnexpectedMulticallOutput    if the decoded argument count is invalid
//   - ErrUnexpectedMulticallOutputType if the decoded type is not [][]byte
func DecodeMulticall(encodedCalldata []byte) ([][]byte, error) {
	// Ensure calldata contains at least the 4-byte function selector.
	if len(encodedCalldata) < 4 {
		return nil, ErrShortCalldata
	}

	// Strip the 4-byte function selector.
	data := encodedCalldata[4:]

	// Decode the ABI-encoded input arguments for the `multicall` method.
	out, err := MulticallABI.Methods["multicall"].Inputs.Unpack(data)
	if err != nil {
		return nil, err
	}

	// The multicall method is expected to have exactly one argument: bytes[].
	if len(out) != 1 {
		return nil, errors.Wrapf(ErrUnexpectedMulticallOutput, "length: %d", len(out))
	}

	// Type assertion to [][]byte, which corresponds to ABI type bytes[].
	result, ok := out[0].([][]byte)
	if !ok {
		return nil, ErrUnexpectedMulticallOutputType
	}

	return result, nil
}

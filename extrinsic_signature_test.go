package scalecodec

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
)

func signatureTestMetadata() *types.MetadataStruct {
	return &types.MetadataStruct{
		MetadataVersion: 14,
		Extrinsic: &types.ExtrinsicMetadata{SignedExtensions: []types.SignedExtensions{
			{Identifier: "CheckMortality", TypeString: "EraExtrinsic"},
			{Identifier: "CheckNonce", TypeString: "Compact<U32>"},
			{Identifier: "ChargeTransactionPayment", TypeString: "Compact<U128>"},
		}},
		CallIndex: map[string]types.CallIndex{"0000": {
			Module: types.MetadataModules{Name: "System"},
			Call:   types.MetadataCalls{Name: "remark"},
		}},
	}
}

func signatureTestBody(signature []byte, version byte) []byte {
	body := append([]byte{version}, bytes.Repeat([]byte{0xee}, 32)...)
	body = append(body, signature...)
	if version == 0x85 {
		body = append(body, 0) // existing signed-V5 extension-version format
	}
	// Immortal era, nonce 3, tip 5, followed by System.remark with no fields.
	return append(body, 0, 12, 20, 0, 0)
}

func signatureTestDecode(t *testing.T, body []byte, metadata *types.MetadataStruct) *GenericExtrinsic {
	t.Helper()
	if body[0] == 0x85 {
		// The existing signed-V5 path decodes nonce before its extension loop.
		copyMetadata := *metadata
		copyExtrinsic := *metadata.Extrinsic
		copyExtrinsic.SignedExtensions = []types.SignedExtensions{{Identifier: "ChargeTransactionPayment", TypeString: "Compact<U128>"}}
		copyMetadata.Extrinsic = &copyExtrinsic
		metadata = &copyMetadata
	}
	prefix := utiles.HexToBytes(types.Encode("Compact<U32>", len(body)))
	input := append(prefix, body...)
	decoder := ExtrinsicDecoder{}
	decoder.Init(scaleBytes.ScaleBytes{Data: input}, &types.ScaleDecoderOption{Metadata: metadata})
	decoder.Process()
	require.Equal(t, len(input), decoder.Data.Offset)
	result := decoder.Value.(*GenericExtrinsic)
	require.Equal(t, "remark", result.CallModuleFunction)
	require.Equal(t, 3, result.Nonce)
	require.Equal(t, "5", result.Tip.String())
	require.Equal(t, "00", result.Era)
	return result
}

func TestCompositeSignatureSpan(t *testing.T) {
	t.Cleanup(types.Clean)
	for _, test := range []struct {
		name    string
		typeDef source.TypeStruct
		encoded []byte
	}{
		{"tuple", source.TypeStruct{Type: "string", TypeString: "([u8; 3], [u8; 2])"}, []byte{0xab, 0xcd, 0xef, 0xde, 0xad}},
		{"enum with composite payload", source.TypeStruct{Type: "enum", TypeMapping: [][]string{{"Composite", "([u8; 3], [u8; 2])"}}}, []byte{0, 0xab, 0xcd, 0xef, 0xde, 0xad}},
	} {
		t.Run(test.name, func(t *testing.T) {
			types.Clean()
			types.RegCustomTypes(map[string]source.TypeStruct{
				"ExtrinsicSigner":    {Type: "string", TypeString: "AccountId"},
				"ExtrinsicSignature": test.typeDef,
			})
			result := signatureTestDecode(t, signatureTestBody(test.encoded, 0x84), signatureTestMetadata())
			require.Equal(t, hex.EncodeToString(test.encoded), result.Signature)
			require.Equal(t, test.encoded, utiles.HexToBytes(types.Encode("ExtrinsicSignature", result.SignatureRaw)))
		})
	}
	for _, version := range []byte{0x84, 0x85} {
		for _, first := range []byte{0, 1, 2} {
			t.Run(fmt.Sprintf("version-%x/first-%d", version, first), func(t *testing.T) {
				types.Clean()
				types.RegCustomTypes(map[string]source.TypeStruct{
					"ExtrinsicSigner": {Type: "string", TypeString: "AccountId"},
					"ExtrinsicSignature": {Type: "struct", TypeMapping: [][]string{
						{"public", "[u8; 3]"}, {"signature", "([u8; 2], [u8; 3])"},
					}},
				})
				metadata := signatureTestMetadata()
				encoded := []byte{first, 0xab, 0xcd, 0xde, 0xad, 0xbe, 0xef, 0xff}
				var firstRaw interface{}
				for repeat := 0; repeat < 20; repeat++ {
					result := signatureTestDecode(t, signatureTestBody(encoded, version), metadata)
					require.Equal(t, hex.EncodeToString(encoded), result.Signature)
					require.IsType(t, map[string]interface{}{}, result.SignatureRaw)
					if repeat == 0 {
						firstRaw = result.SignatureRaw
					}
					require.Equal(t, firstRaw, result.SignatureRaw)
					require.Equal(t, encoded, utiles.HexToBytes(types.EncodeWithOpt("ExtrinsicSignature", result.SignatureRaw, &types.ScaleDecoderOption{Metadata: metadata})))
				}
				// Recompute the body length so rejection must come from decoding,
				// rather than a mismatched outer length. Cut within the signature.
				body := signatureTestBody(encoded, version)
				for end := 33; end < 33+len(encoded); end++ {
					prefix := utiles.HexToBytes(types.Encode("Compact<U32>", end))
					input := append(prefix, body[:end]...)
					decoder := ExtrinsicDecoder{}
					decoder.Init(scaleBytes.ScaleBytes{Data: input}, &types.ScaleDecoderOption{Metadata: metadata})
					require.Panics(t, decoder.Process, "truncated signature at %d", end-33)
				}
			})
		}
	}
}

func TestSignatureOutputConventions(t *testing.T) {
	t.Cleanup(types.Clean)
	for _, version := range []byte{0x84, 0x85} {
		for variant, name := range []string{"Ed25519", "Sr25519", "Ecdsa"} {
			t.Run(fmt.Sprintf("version-%x/%s", version, name), func(t *testing.T) {
				types.Clean()
				types.RegCustomTypes(map[string]source.TypeStruct{
					"ExtrinsicSigner": {Type: "string", TypeString: "AccountId"},
					"ExtrinsicSignature": {Type: "enum", TypeMapping: [][]string{
						{"Ed25519", "[u8; 64]"}, {"Sr25519", "[u8; 64]"}, {"Ecdsa", "[u8; 65]"},
					}},
				})
				length := 64
				if variant == 2 {
					length = 65
				}
				payload := bytes.Repeat([]byte{0xee}, length)
				result := signatureTestDecode(t, signatureTestBody(append([]byte{byte(variant)}, payload...), version), signatureTestMetadata())
				// Keep the established enum payload output, including its 0x prefix.
				require.Equal(t, "0x"+hex.EncodeToString(payload), result.Signature)
				require.Equal(t, map[string]interface{}{name: result.Signature}, result.SignatureRaw)
			})
		}
	}
	t.Run("direct string signature", func(t *testing.T) {
		types.Clean()
		types.RegCustomTypes(map[string]source.TypeStruct{
			"ExtrinsicSigner":    {Type: "string", TypeString: "AccountId"},
			"ExtrinsicSignature": {Type: "string", TypeString: "[u8; 4]"},
		})
		result := signatureTestDecode(t, signatureTestBody([]byte{0xde, 0xad, 0xbe, 0xef}, 0x84), signatureTestMetadata())
		require.Equal(t, "0xdeadbeef", result.Signature)
		require.Equal(t, result.Signature, result.SignatureRaw)
	})
	t.Run("one field struct is not an enum", func(t *testing.T) {
		types.Clean()
		types.RegCustomTypes(map[string]source.TypeStruct{
			"ExtrinsicSigner":    {Type: "string", TypeString: "AccountId"},
			"ExtrinsicSignature": {Type: "struct", TypeMapping: [][]string{{"Sr25519", "Bytes"}}},
		})
		encoded := []byte{8, 0xde, 0xad} // Vec length is part of the composite envelope.
		result := signatureTestDecode(t, signatureTestBody(encoded, 0x84), signatureTestMetadata())
		require.Equal(t, "08dead", result.Signature)
		require.Equal(t, map[string]interface{}{"Sr25519": "0xdead"}, result.SignatureRaw)
	})
}

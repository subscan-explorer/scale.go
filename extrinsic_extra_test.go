package scalecodec

import (
	"testing"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
)

func TestDecodeExtrinsicExtraUsesMetadataTypes(t *testing.T) {
	// Simulate another metadata instance having registered a different
	// ExtrinsicExtra layout. Decoding must use the current metadata's signed
	// extension types instead of this process-global entry.
	types.RegCustomTypes(map[string]source.TypeStruct{
		"ExtrinsicExtra": {
			Type:        "struct",
			TypeMapping: [][]string{{"col1", "u32"}},
		},
	})

	metadata := types.MetadataStruct{
		MetadataVersion: 14,
		Extrinsic: &types.ExtrinsicMetadata{
			SignedExtensions: []types.SignedExtensions{
				{Identifier: "First", TypeString: "u8"},
				{Identifier: "Empty", TypeString: "NULL"},
				{Identifier: "Second", TypeString: "u16"},
			},
		},
	}

	decoder := ExtrinsicDecoder{}
	decoder.Init(
		scaleBytes.ScaleBytes{Data: []byte{7, 0x09, 0x08}},
		&types.ScaleDecoderOption{Metadata: &metadata},
	)

	result := GenericExtrinsic{}
	if !decoder.decodeExtrinsicExtra(&result) {
		t.Fatal("expected v14 signed extensions to be decoded")
	}
	if decoder.Data.Offset != 3 {
		t.Fatalf("expected all extension bytes to be consumed, got offset %d", decoder.Data.Offset)
	}
	if got := result.SignedExtensions["First"]; got != 7 {
		t.Fatalf("unexpected First extension value: %v", got)
	}
	if got := result.SignedExtensions["Second"]; got != uint16(0x0809) {
		t.Fatalf("unexpected Second extension value: %v", got)
	}
}

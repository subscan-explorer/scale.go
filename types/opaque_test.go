package types

import (
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
	"testing"
)

func opaqueTestMetadata() *MetadataStruct {
	return &MetadataStruct{CallIndex: map[string]CallIndex{
		"0000": {Module: MetadataModules{Name: "Example"}, Call: MetadataCalls{Name: "value", Args: []MetadataModuleCallArgument{{Name: "value", Type: "U32"}}}},
		"0001": {Module: MetadataModules{Name: "Example"}, Call: MetadataCalls{Name: "nested", Args: []MetadataModuleCallArgument{{Name: "call", Type: "Call"}, {Name: "weight", Type: "U64"}}}},
		"4142": {Module: MetadataModules{Name: "Example"}, Call: MetadataCalls{Name: "ascii_index"}},
	}}
}

func TestOpaquePayloadBoundary(t *testing.T) {
	for _, typ := range []string{"OpaqueCall", "WrapperOpaque<Call>"} {
		t.Run(typ, func(t *testing.T) {
			for _, tc := range []struct{ name, raw, want string }{
				{"empty", "00", "0x"},
				{"unknown call", "086902", "0x6902"},
				{"inner underflow", "1000000102", "0x00000102"},
				{"inner trailing bytes", "1c000078563412ff", "0x000078563412ff"},
				{"partial nested call", "200001000078563412", "0x0001000078563412"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					input := utiles.HexToBytes(tc.raw + "010807060504030201")
					e := ScaleDecoder{}
					e.Init(scaleBytes.ScaleBytes{Data: input}, &ScaleDecoderOption{Metadata: opaqueTestMetadata()})
					require.NotPanics(t, func() { require.Equal(t, tc.want, e.ProcessAndUpdateData(typ)) })
					require.Equal(t, len(utiles.HexToBytes(tc.raw)), e.Data.Offset)
					require.Empty(t, e.InternalCall, "failed inner decodes must not publish partial calls")
					require.Equal(t, true, e.ProcessAndUpdateData("Bool"))
					require.Equal(t, uint64(0x0102030405060708), e.ProcessAndUpdateData("U64"))
					require.Equal(t, len(input), e.Data.Offset)
				})
			}
			for _, raw := range []string{"", "01", "020000", "03ffffff", "18000001", "070000000001"} {
				t.Run("truncated or invalid envelope "+raw, func(t *testing.T) {
					e := ScaleDecoder{}
					e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, &ScaleDecoderOption{Metadata: opaqueTestMetadata()})
					require.Panics(t, func() { e.ProcessAndUpdateData(typ) })
				})
			}
			for _, tc := range []struct{ raw, name, call string }{
				{"18000078563412", "value", "000078563412"},
				{"084142", "ascii_index", "4142"},
			} {
				t.Run("valid "+tc.name, func(t *testing.T) {
					e := ScaleDecoder{}
					e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(tc.raw)}, &ScaleDecoderOption{Metadata: opaqueTestMetadata()})
					v := e.ProcessAndUpdateData(typ).(map[string]interface{})
					require.Equal(t, tc.name, v["call_name"])
					require.Equal(t, []string{tc.call}, e.InternalCall)
					require.Equal(t, len(utiles.HexToBytes(tc.raw)), e.Data.Offset)
				})
			}
		})
	}
}

func TestWrapperOpaqueNonCall(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want interface{}
	}{
		{"1078563412", uint32(0x12345678)},
		{"087856", "0x7856"},
		{"1478563412ff", "0x78563412ff"},
	} {
		e := ScaleDecoder{}
		e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(tc.raw + "01")}, nil)
		require.Equal(t, tc.want, e.ProcessAndUpdateData("WrapperOpaque<U32>"))
		require.Equal(t, true, e.ProcessAndUpdateData("Bool"))
	}
}

func TestOpaqueValidNestedCall(t *testing.T) {
	for _, typ := range []string{"OpaqueCall", "WrapperOpaque<Call>"} {
		e := ScaleDecoder{}
		raw := "4000010000785634120807060504030201"
		e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, &ScaleDecoderOption{Metadata: opaqueTestMetadata()})
		value := e.ProcessAndUpdateData(typ).(map[string]interface{})
		require.Equal(t, "nested", value["call_name"])
		require.Equal(t, []string{"000078563412", "00010000785634120807060504030201"}, e.InternalCall)
		require.Equal(t, len(utiles.HexToBytes(raw)), e.Data.Offset)
	}
}

func TestWrapperOpaqueZeroSizedType(t *testing.T) {
	e := ScaleDecoder{}
	e.Init(scaleBytes.ScaleBytes{Data: []byte{0, 1}}, nil)
	require.Nil(t, e.ProcessAndUpdateData("WrapperOpaque<Null>"))
	require.Equal(t, true, e.ProcessAndUpdateData("Bool"))
}

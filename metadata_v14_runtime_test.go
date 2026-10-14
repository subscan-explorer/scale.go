package scalecodec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
)

// One portable type (id 0, primitive U8), no pallets, extrinsic type 0/version 4,
// no signed extensions. The final RuntimeMetadataV14.ty follows this prefix.
var v14RuntimePrefix = []byte{'m', 'e', 't', 'a', 14, 4, 0, 0, 0, 5, 3, 0, 0, 0, 4, 0}

func TestMetadataV14RuntimeType(t *testing.T) {
	t.Cleanup(types.Clean)
	t.Run("retain zero ID and consume exactly", func(t *testing.T) {
		types.Clean()
		input := append(append([]byte{}, v14RuntimePrefix...), 0)
		decoder := MetadataDecoder{}
		decoder.Init(input)
		require.NoError(t, decoder.Process())
		require.NotNil(t, decoder.Metadata.Type)
		require.Equal(t, 0, *decoder.Metadata.Type)
		require.Equal(t, len(input), decoder.Data.Offset)
	})
	for _, test := range []struct {
		name string
		tail []byte
	}{
		{"missing type", nil},
		{"truncated two byte type", []byte{1}},
		{"truncated four byte type", []byte{2, 0, 0}},
		{"truncated large mode type", []byte{3, 0, 0, 0}},
		{"unknown type", []byte{4}}, // compact ID 1 is not in this registry
	} {
		t.Run(test.name, func(t *testing.T) {
			types.Clean()
			decoder := MetadataDecoder{}
			decoder.Init(append(append([]byte{}, v14RuntimePrefix...), test.tail...))
			require.Panics(t, func() { _ = decoder.Process() })
		})
	}
	t.Run("reject trailing bytes", func(t *testing.T) {
		types.Clean()
		decoder := MetadataDecoder{}
		decoder.Init(append(append([]byte{}, v14RuntimePrefix...), 0, 0xff))
		err := decoder.Process()
		require.Error(t, err)
		if err != nil {
			require.Contains(t, err.Error(), "trailing bytes")
		}
	})
}

func TestMetadataV14QuipRuntimeType(t *testing.T) {
	t.Cleanup(types.Clean)
	const dir = "testdata/quip-v14"
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		return data
	}
	var provenance struct {
		Calls []struct {
			File, Method, RequestSHA256, ResponseSHA256 string
			Params                                      []string
		}
	}
	provenanceBytes := read("provenance.json")
	require.Equal(t, "6474b710ab8e3ae11f4bc5f3379df068083bd044e3da8a63660b8d4a225fad2b", signatureFixtureSHA(provenanceBytes))
	require.NoError(t, json.Unmarshal(provenanceBytes, &provenance))
	require.Len(t, provenance.Calls, 2)
	for index, call := range provenance.Calls {
		t.Run(fmt.Sprintf("runtime-%d", 117+index), func(t *testing.T) {
			types.Clean()
			requestBytes := read(call.File + ".request.json")
			responseBytes := read(call.File + ".response.json")
			require.Equal(t, call.RequestSHA256, signatureFixtureSHA(requestBytes))
			require.Equal(t, call.ResponseSHA256, signatureFixtureSHA(responseBytes))
			var request struct {
				Method string
				Params []string
			}
			require.NoError(t, json.Unmarshal(requestBytes, &request))
			require.Equal(t, "state_call", request.Method)
			require.Equal(t, call.Params, request.Params)
			require.Equal(t, []string{"Metadata_metadata_at_version", "0x0e000000"}, request.Params[:2])
			var response struct{ Result string }
			require.NoError(t, json.Unmarshal(responseBytes, &response))
			input := utiles.HexToBytes(response.Result)
			decoder := MetadataDecoder{}
			decoder.Init(input) // captured Option<OpaqueMetadata> wrapper retained
			decoder.Spec = 117 + index
			require.NoError(t, decoder.Process())
			require.Equal(t, 14, decoder.Metadata.MetadataVersion)
			require.Equal(t, len(input), decoder.Data.Offset)
			require.NotNil(t, decoder.Metadata.Type)
			expectedID := []int{273, 274}[index]
			require.Equal(t, expectedID, *decoder.Metadata.Type)
			registryDecoder := MetadataDecoder{}
			registryDecoder.Init(input)
			registry, err := registryDecoder.PortableRegistry()
			require.NoError(t, err)
			require.Equal(t, []string{"quip_protocol_runtime", "Runtime"}, registry[expectedID].Path)
			encodedID := utiles.HexToBytes(types.Encode("SiLookupTypeId", expectedID))
			require.Equal(t, encodedID, input[len(input)-len(encodedID):])
			invalidID := 100000
			_, exists := registry[invalidID]
			require.False(t, exists)
			invalid := append([]byte{}, input[:len(input)-len(encodedID)]...)
			invalid = append(invalid, utiles.HexToBytes(types.Encode("SiLookupTypeId", invalidID))...)
			invalidDecoder := MetadataDecoder{}
			invalidDecoder.Init(invalid)
			require.PanicsWithValue(t, "metadata V14 runtime type 100000 not found in portable registry", func() { _ = invalidDecoder.Process() })
			trailing := MetadataDecoder{}
			trailing.Init(append(append([]byte{}, input...), 0xff))
			require.EqualError(t, trailing.Process(), "metadata V14 has 1 trailing bytes")
			for cut := 1; cut <= len(encodedID); cut++ {
				truncated := MetadataDecoder{}
				truncated.Init(input[:len(input)-cut])
				require.Panics(t, func() { _ = truncated.Process() })
			}
		})
	}
}

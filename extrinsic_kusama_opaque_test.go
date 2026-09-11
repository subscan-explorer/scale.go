package scalecodec_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	scale "github.com/itering/scale.go"
	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
)

func TestKusamaOpaqueHistoricalBlocks(t *testing.T) {
	// Type registration is process-global. Keep historical chain definitions out
	// of unrelated network tests and exercise each runtime in a fresh process.
	specText := os.Getenv("SCALE_KUSAMA_OPAQUE_SPEC")
	if specText == "" {
		for _, spec := range []int{2015, 2029, 9070, 9260} {
			t.Run(strconv.Itoa(spec), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestKusamaOpaqueHistoricalBlocks$", "-test.v")
				cmd.Env = append(os.Environ(), fmt.Sprintf("SCALE_KUSAMA_OPAQUE_SPEC=%d", spec))
				out, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", out)
			})
		}
		return
	}
	spec, err := strconv.Atoi(specText)
	require.NoError(t, err)
	raw, err := os.ReadFile("testdata/kusama_opaque/types.json")
	require.NoError(t, err)
	types.RegCustomTypes(source.LoadTypeRegistry(raw))
	raw, err = os.ReadFile(fmt.Sprintf("testdata/kusama_opaque/metadata-%d.scale", spec))
	require.NoError(t, err)
	m := scale.MetadataDecoder{}
	m.Init(raw)
	m.Spec = spec
	require.NoError(t, m.Process())
	fixture, err := os.ReadFile("testdata/kusama_opaque/blocks.json")
	require.NoError(t, err)
	var blocks []struct {
		Height, Spec    int
		MetadataVersion int `json:"metadata_version"`
		Extrinsics      []struct {
			Opaque       bool
			Raw          string
			ExpectedCall string `json:"expected_call"`
			Params       json.RawMessage
		}
		Events     string
		EventCount int `json:"event_count"`
		Logs       []string
	}
	require.NoError(t, json.Unmarshal(fixture, &blocks))
	require.Len(t, blocks, 16)
	count := 0
	for _, block := range blocks {
		if block.Spec != spec {
			continue
		}
		count++
		t.Run(strconv.Itoa(block.Height), func(t *testing.T) {
			require.Equal(t, block.MetadataVersion, m.Metadata.MetadataVersion)
			opt := types.ScaleDecoderOption{Metadata: &m.Metadata, Spec: spec}
			for i, x := range block.Extrinsics {
				t.Run(fmt.Sprintf("extrinsic_%d", i), func(t *testing.T) {
					input := utiles.HexToBytes(x.Raw)
					e := scale.ExtrinsicDecoder{}
					e.Init(scaleBytes.ScaleBytes{Data: input}, &opt)
					require.NotPanics(t, e.Process)
					require.Equal(t, len(input), e.Data.Offset)
					v := e.Value.(*scale.GenericExtrinsic)
					require.Equal(t, x.ExpectedCall, v.CallModule+"."+v.CallModuleFunction)
					if x.Opaque {
						box := types.BoxCall{CallIndex: v.CallCode}
						for _, param := range v.Params {
							box.Params = append(box.Params, types.ExtrinsicParam{Name: param.Name, Type: param.Type, Value: param.Value})
						}
						require.Equal(t, utiles.TrimHex(v.CallCode)+utiles.TrimHex(v.ParamsRaw), types.EncodeWithOpt("Call", box, &opt), "re-encode the exact outer call including opaque nested arguments")

					}

					got, err := json.Marshal(v.Params)
					require.NoError(t, err)
					var gotValue, wantValue interface{}
					require.NoError(t, json.Unmarshal(got, &gotValue))
					require.NoError(t, json.Unmarshal(x.Params, &wantValue))
					require.Equal(t, opaqueParamValues(wantValue), opaqueParamValues(gotValue))
				})
			}
			input := utiles.HexToBytes(block.Events)
			d := types.ScaleDecoder{}
			d.Init(scaleBytes.ScaleBytes{Data: input}, &opt)
			require.Equal(t, block.EventCount, d.ProcessAndUpdateData("Compact<u32>"))
			event := scale.EventRecord{Metadata: &m.Metadata}
			event.Init(d.Data, &opt)
			for i := 0; i < block.EventCount; i++ {
				require.NotPanics(t, func() { event.Process() })
			}
			require.Equal(t, len(input), event.Data.Offset)
			for _, log := range block.Logs {
				input := utiles.HexToBytes(log)
				d := types.ScaleDecoder{}
				d.Init(scaleBytes.ScaleBytes{Data: input}, nil)
				require.NotPanics(t, func() { d.ProcessAndUpdateData("LogDigest") })
				require.Equal(t, len(input), d.Data.Offset)
			}
		})
	}
	require.Equal(t, map[int]int{2015: 10, 2029: 2, 9070: 1, 9260: 3}[spec], count)
}

// Expected values were captured by reading only the opaque argument as bytes.
// Its diagnostic type label differs; compare parameter names and values while
// retaining every byte (including the canonical empty opaque value "0x").
func opaqueParamValues(v interface{}) interface{} {
	switch value := v.(type) {
	case []interface{}:
		for i, item := range value {
			value[i] = opaqueParamValues(item)
		}
	case map[string]interface{}:
		if value["type"] == "HexBytes" && value["value"] == "" {
			value["value"] = "0x"
		}
		delete(value, "type")
		delete(value, "type_name")
		for k, item := range value {
			value[k] = opaqueParamValues(item)
		}
	}
	return v
}

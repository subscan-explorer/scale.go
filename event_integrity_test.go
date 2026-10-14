package scalecodec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/stretchr/testify/require"
)

func TestEventsConsumeWholeVector(t *testing.T) {
	m := &types.MetadataStruct{MetadataVersion: 14, EventIndex: map[string]types.EventIndex{"0000": {Module: types.MetadataModules{Name: "Example"}, Call: types.MetadataEvents{Name: "Changed", Args: []string{"U32"}}}}}
	// ApplyExtrinsic(2), Example.Changed(7), one 32-byte topic.
	record := []byte{0, 2, 0, 0, 0, 0, 0, 7, 0, 0, 0, 4}
	record = append(record, bytes.Repeat([]byte{0xab}, 32)...)
	valid := append([]byte{4}, record...)
	decode := func(raw []byte) *EventsDecoder {
		e := &EventsDecoder{}
		e.Init(scaleBytes.ScaleBytes{Data: raw}, &types.ScaleDecoderOption{Metadata: m})
		e.Process()
		return e
	}
	for _, raw := range [][]byte{{0}, valid, append(append([]byte{8}, record...), record...), append([]byte{1, 1}, bytes.Repeat(record, 64)...)} {
		e := decode(raw)
		require.Equal(t, len(raw), e.Data.Offset)
		require.Zero(t, e.Data.GetRemainingLength())
	}
	e := decode(valid)
	events := e.Value.([]interface{})
	require.Len(t, events, 1)
	event := events[0].(map[string]interface{})
	require.Equal(t, 2, event["extrinsic_idx"])
	require.Equal(t, 0, event["event_idx"])
	require.Equal(t, "07000000", event["params_raw"])
	require.Len(t, event["topic"], 1)
	// Every proper prefix is incomplete, including topic hashes the old decoder
	// padded silently. No partial result may escape a failed Process call.
	for size := 0; size < len(valid); size++ {
		t.Run(fmt.Sprint("truncated/", size), func(t *testing.T) {
			decoder := &EventsDecoder{}
			decoder.Init(scaleBytes.ScaleBytes{Data: valid[:size]}, &types.ScaleDecoderOption{Metadata: m})
			require.Panics(t, decoder.Process)
			require.Nil(t, decoder.Value)
		})
	}
	for _, extra := range [][]byte{{0}, {1, 2, 3}} {
		require.PanicsWithValue(t, fmt.Sprintf("events have %d trailing bytes", len(extra)), func() { decode(append(append([]byte{}, valid...), extra...)) })
	}
	require.Panics(t, func() { decode([]byte{0, 0}) })
	for _, prefix := range [][]byte{{1, 0}, {2, 0, 0, 0}, {3, 0, 0, 0, 0}, {7, 0, 0, 0, 0, 0}} {
		require.Panics(t, func() { decode(prefix) })
	}
	invalidPhase := append([]byte{}, valid...)
	invalidPhase[1] = 3
	require.Panics(t, func() { decode(invalidPhase) })
	noTopicCount := append([]byte{4}, record[:11]...)
	require.Panics(t, func() { decode(noTopicCount) })
	for _, phase := range []byte{1, 2} {
		raw := []byte{4, phase, 0, 0, 9, 0, 0, 0, 0}
		event := decode(raw).Value.([]interface{})[0].(map[string]interface{})
		require.Equal(t, int(phase), event["phase"])
		require.Equal(t, 0, event["extrinsic_idx"])
	}
}

func TestQuipEventFixtureConsumption(t *testing.T) {
	t.Cleanup(types.Clean)
	total := 0
	for _, fixture := range []struct{ dir, sha string }{
		{"aglais-117-118", "a77d59562ecab6aaf80173e70a1f6dcf210815f8af471c7ce356e7ef08bd4b89"},
		{"local-119", "eab61fd6227b41dadc03de4df98ebc2c4d176506cb4c69124fa76b4ea7be4e48"},
	} {
		dir := filepath.Join("testdata", "quip-metadata-conformance", fixture.dir)
		rawManifest := signatureFixtureRead(t, filepath.Join(dir, "manifest.json"))
		require.Equal(t, fixture.sha, signatureFixtureSHA(rawManifest))
		var manifest struct {
			Blocks []struct {
				Number, ExecutionSpecVersion int
				MetadataFile, EventsFile     string
			}
			Files map[string]struct{ SHA256 string }
		}
		require.NoError(t, json.Unmarshal(rawManifest, &manifest))
		read := func(name string) []byte {
			raw := signatureFixtureRead(t, filepath.Join(dir, name))
			require.Equal(t, manifest.Files[name].SHA256, signatureFixtureSHA(raw))
			return raw
		}
		for _, block := range manifest.Blocks {
			t.Run(fmt.Sprintf("%s/%d", fixture.dir, block.Number), func(t *testing.T) {
				types.Clean()
				metadata := MetadataDecoder{}
				metadata.Init(read(block.MetadataFile))
				metadata.Spec = block.ExecutionSpecVersion
				require.NoError(t, metadata.Process())
				raw := read(block.EventsFile)
				decode := func(data []byte) *EventsDecoder {
					e := &EventsDecoder{}
					e.Init(scaleBytes.ScaleBytes{Data: data}, &types.ScaleDecoderOption{Metadata: &metadata.Metadata, Spec: block.ExecutionSpecVersion})
					e.Process()
					return e
				}
				first := decode(raw)
				require.Equal(t, len(raw), first.Data.Offset)
				require.Zero(t, first.Data.GetRemainingLength())
				events := first.Value.([]interface{})
				total += len(events)
				require.Equal(t, first.Value, decode(raw).Value, "repeat decode must be stable")
				for _, data := range [][]byte{append(append([]byte{}, raw...), 0), raw[:len(raw)-1]} {
					require.Panics(t, func() { decode(data) })
				}
			})
		}
	}
	require.Equal(t, 107, total)
}

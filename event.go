package scalecodec

import (
	"encoding/binary"
	"fmt"

	scaleType "github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type EventsDecoder struct {
	scaleType.Vec
	Metadata *scaleType.MetadataStruct
}

func (e *EventsDecoder) Init(data scaleBytes.ScaleBytes, option *scaleType.ScaleDecoderOption) {
	if option == nil || option.Metadata == nil {
		panic("EventsDecoder option metadata required")
	}
	e.TypeString = "Vec<EventRecord>"
	e.Metadata = option.Metadata
	e.Vec.Init(data, option)
}

type EventParam struct {
	Type     string      `json:"type"`
	Name     string      `json:"name,omitempty"`
	TypeName string      `json:"type_name"`
	Value    interface{} `json:"value"`
}

func (e *EventsDecoder) Process() {
	e.Value = nil
	count, prefix := eventCompactCount(e.Data.Data[e.Data.Offset:])
	minimumRecord := 3 // phase + pallet/event indices
	if e.Metadata.MetadataVersion >= 5 {
		minimumRecord++ // topics vector count
	}
	if uint64(count)*uint64(minimumRecord) > uint64(e.Data.GetRemainingLength()-prefix) {
		panic("event count exceeds remaining record bytes")
	}
	elementCount := e.ProcessAndUpdateData("Compact<u32>").(int)

	er := EventRecord{Metadata: e.Metadata}
	er.Data = e.Data
	er.Spec = e.Spec
	er.ScaleDecoder.Metadata = e.Metadata
	var result []interface{}
	for i := 0; i < elementCount; i++ {
		element := er.Process()
		element["event_idx"] = i
		result = append(result, element)
	}
	// EventRecord owns a copy of ScaleBytes. Publish its final cursor before
	// accepting the root vector; previously only the count prefix was consumed.
	e.Data = er.Data
	if e.Data.GetRemainingLength() != 0 {
		panic(fmt.Sprintf("events have %d trailing bytes", e.Data.GetRemainingLength()))
	}
	e.Value = result
}

// Validate counts before invoking the legacy compact/vector decoders, which
// otherwise accept short/noncanonical prefixes and can pad truncated topics.
func eventCompactCount(data []byte) (uint32, int) {
	if len(data) == 0 {
		panic("missing event vector count")
	}
	switch data[0] & 3 {
	case 0:
		return uint32(data[0] >> 2), 1
	case 1:
		if len(data) < 2 {
			panic("truncated event vector count")
		}
		count := uint32(binary.LittleEndian.Uint16(data) >> 2)
		if count < 64 {
			panic("noncanonical event vector count")
		}
		return count, 2
	case 2:
		if len(data) < 4 {
			panic("truncated event vector count")
		}
		count := binary.LittleEndian.Uint32(data) >> 2
		if count < 1<<14 {
			panic("noncanonical event vector count")
		}
		return count, 4
	default:
		if data[0] != 3 || len(data) < 5 {
			panic("invalid u32 event vector count")
		}
		count := binary.LittleEndian.Uint32(data[1:])
		if count < 1<<30 {
			panic("noncanonical event vector count")
		}
		return count, 5
	}
}

type EventRecord struct {
	scaleType.ScaleDecoder
	Metadata     *scaleType.MetadataStruct
	Phase        int                      `json:"phase"`
	ExtrinsicIdx int                      `json:"extrinsic_idx"`
	Type         string                   `json:"type"`
	Params       []EventParam             `json:"params"`
	Event        scaleType.MetadataEvents `json:"event"`
	Topic        []string                 `json:"topic"`
	ParamsRaw    string                   `json:"params_raw"`
}

func (e *EventRecord) Process() map[string]interface{} {
	e.Params = []EventParam{}
	e.ExtrinsicIdx = 0
	e.Topic = []string{}

	if e.Data.GetRemainingLength() < 3 {
		panic("truncated event record")
	}
	e.Phase = e.GetNextU8()
	if e.Phase > 2 {
		panic("invalid event phase")
	}
	// enum
	// isApplyExtrinsic: bool;
	// asApplyExtrinsic: u32;
	// isFinalization: boolean;
	// isInitialization: boolean;

	if e.Phase == 0 {
		if e.Data.GetRemainingLength() < 6 {
			panic("truncated apply-extrinsic event")
		}
		e.ExtrinsicIdx = int(e.ProcessAndUpdateData("U32").(uint32))
	}
	e.Type = utiles.BytesToHex(e.NextBytes(2))

	call, ok := e.Metadata.EventIndex[e.Type]
	if !ok {
		panic(fmt.Sprintf("Not find Event Lookup %s, please check metadata info", e.Type))
	}

	e.Event = call.Call
	e.Module = call.Module.Name
	offset := e.Data.Offset
	for index, argType := range e.Event.Args {
		value := e.ProcessAndUpdateData(argType)
		param := EventParam{Type: argType, Value: value}
		if len(e.Event.ArgsTypeName) == len(e.Event.Args) {
			param.TypeName = e.Event.ArgsTypeName[index]
		}
		if len(e.Event.ArgsName) == len(e.Event.Args) {
			param.Name = e.Event.ArgsName[index]
		}
		e.Params = append(e.Params, param)
	}
	e.ParamsRaw = utiles.BytesToHex(e.Data.Data[offset:e.Data.Offset])

	if e.Metadata.MetadataVersion >= 5 {
		count, prefix := eventCompactCount(e.Data.Data[e.Data.Offset:])
		if uint64(count)*32 > uint64(e.Data.GetRemainingLength()-prefix) {
			panic("truncated event topics")
		}
		if topic := e.ProcessAndUpdateData("Vec<Hash>"); topic != nil {
			topicValue := topic.([]interface{})
			for _, v := range topicValue {
				e.Topic = append(e.Topic, v.(string))
			}
		}
	}

	return map[string]interface{}{
		"phase":         e.Phase,
		"extrinsic_idx": e.ExtrinsicIdx,
		"type":          e.Type,
		"module_id":     call.Module.Name,
		"event_id":      e.Event.Name,
		"params":        e.Params,
		"topic":         e.Topic,
		"params_raw":    e.ParamsRaw,
	}

}

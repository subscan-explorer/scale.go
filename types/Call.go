package types

import (
	"encoding/json"
	"fmt"

	"github.com/itering/scale.go/utiles"
)

type Call struct {
	ScaleDecoder
}

func (s *Call) Process() {
	start := s.Data.Offset
	callIndex := utiles.BytesToHex(s.NextBytes(2))
	callModule, ok := s.Metadata.CallIndex[callIndex]
	if !ok {
		panic(fmt.Sprintf("Not find Extrinsic Lookup %s, please check metadata info", callIndex))
	}
	result := map[string]interface{}{
		"call_index":  callIndex,
		"call_name":   callModule.Call.Name,
		"call_module": callModule.Module.Name,
	}
	var param []ExtrinsicParam
	s.Module = callModule.Module.Name
	for _, arg := range callModule.Call.Args {
		param = append(param, ExtrinsicParam{Name: arg.Name, Type: arg.Type, Value: s.ProcessAndUpdateData(arg.Type)})
	}
	result["params"] = param
	s.InternalCall = append(s.InternalCall, utiles.BytesToHex(s.Data.Data[start:s.Data.Offset]))
	s.Value = result
}

type BoxCall struct {
	CallIndex string           `json:"call_Index"`
	Params    []ExtrinsicParam `json:"params"`
	CallArgs  []ExtrinsicParam `json:"call_args"`
}

func (s *Call) Encode(value interface{}) string {
	if s.Metadata == nil {
		panic("nil metadata")
	}
	var boxCall BoxCall
	switch v := value.(type) {
	case BoxCall:
		boxCall = v
	case map[string]interface{}:
		// Call.Process already supplies typed parameters. A JSON round trip
		// would convert nested u64 values to float64 and corrupt call hashes.
		params, typed := v["params"].([]ExtrinsicParam)
		index, indexed := v["call_index"].(string)
		if typed && indexed {
			boxCall = BoxCall{CallIndex: index, Params: params}
		} else {
			b, err := json.Marshal(v)
			if err != nil {
				panic(fmt.Errorf("marshal boxCall: %w", err))
			}
			if err := json.Unmarshal(b, &boxCall); err != nil {
				panic(fmt.Errorf("input value is not valid boxCall: %w", err))
			}
		}
	default:
		panic("input value is not valid boxCall")
	}
	if len(boxCall.CallArgs) > 0 {
		boxCall.Params = boxCall.CallArgs
	}
	callIndex := utiles.TrimHex(boxCall.CallIndex)
	raw := callIndex
	callModule, ok := s.Metadata.CallIndex[callIndex]
	if !ok {
		panic(fmt.Sprintf("Not find Extrinsic Lookup %s, please check metadata info", callIndex))
	}
	for index, arg := range callModule.Call.Args {
		raw += EncodeWithOpt(arg.Type, boxCall.Params[index].Value, &ScaleDecoderOption{Spec: s.Spec, Metadata: s.Metadata})
	}
	return raw
}

func (s *Call) TypeStructString() string {
	return "Call"
}

type BoxProposal struct {
	Call
}

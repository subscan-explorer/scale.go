package types

import (
	"fmt"
	"strings"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type FixedArray struct {
	ScaleDecoder
	FixedLength int
	SubType     string
}

func (f *FixedArray) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	if option != nil && option.FixedLength != 0 {
		f.FixedLength = option.FixedLength
	}
	f.ScaleDecoder.Init(data, option)
}

func (f *FixedArray) Process() {
	var result []interface{}
	if f.FixedLength > 0 {
		if strings.EqualFold(f.SubType, "u8") {
			value := f.NextBytes(f.FixedLength)
			if utiles.IsASCII(value) {
				f.Value = string(value)
			} else {
				f.Value = utiles.AddHex(utiles.BytesToHex(value))
			}
			return
		}
		for i := 0; i < f.FixedLength; i++ {
			result = append(result, f.ProcessAndUpdateData(f.SubType))
		}
		f.Value = result
	} else {
		f.GetNextU8()
	}
}

func (f *FixedArray) TypeStructString() string {
	return fmt.Sprintf("[%d;%s]", f.FixedLength, getTypeStructString(f.SubType, 0))
}

func (f *FixedArray) Encode(value interface{}) string {
	var raw string
	if valueStr, ok := value.(string); ok {
		if valueStr == "" {
			return ""
		}
		if strings.HasPrefix(valueStr, "0x") {
			return utiles.TrimHex(valueStr)
		} else {
			return utiles.BytesToHex([]byte(valueStr))
		}
	}
	values, ok := asInterfaceSlice(value)
	if ok {
		if len(values) != f.FixedLength {
			panic("fixed length not match")
		}
		subType := f.SubType
		for _, item := range values {
			raw += EncodeWithOpt(subType, item, &ScaleDecoderOption{Spec: f.Spec, Metadata: f.Metadata})
		}
		return raw
	}
	if f.FixedLength == 1 {
		return EncodeWithOpt(f.SubType, value, &ScaleDecoderOption{Spec: f.Spec, Metadata: f.Metadata})
	}
	panic(fmt.Errorf("invalid fixed array input: expected fixed length %d with subtype %q, got value of type %T", f.FixedLength, f.SubType, value))
}

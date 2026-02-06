package types

import (
	"fmt"

	"github.com/itering/scale.go/utiles"
)

type Option struct {
	ScaleDecoder
}

func (o *Option) Process() {
	optionFlag := o.NextBytes(1)
	if o.SubType != "" && utiles.BytesToHex(optionFlag) != "00" {
		if o.SubType == "bool" {
			o.Value = utiles.BytesToHex(optionFlag) == "01"
			return
		}
		o.Value = o.ProcessAndUpdateData(o.SubType)
	}
}

func (o *Option) Encode(value interface{}) string {
	if v, ok := value.(string); ok && v == "" {
		return "00"
	}
	if value == nil {
		return "00"
	}
	switch v := value.(type) {
	case []byte:
		if v == nil {
			return "00"
		}
	case []interface{}:
		if v == nil {
			return "00"
		}
	case map[string]interface{}:
		if v == nil {
			return "00"
		}
	case error:
		if v == nil {
			return "00"
		}
	}
	if o.SubType == "bool" {
		if value.(bool) {
			return "01"
		}
		return "02"
	}
	return "01" + EncodeWithOpt(o.SubType, value, &ScaleDecoderOption{Spec: o.Spec, Metadata: o.Metadata})
}

func (o *Option) TypeStructString() string {
	return fmt.Sprintf("Option<%s>", getTypeStructString(o.SubType, o.RecursiveTime))
}

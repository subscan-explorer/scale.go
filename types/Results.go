package types

import (
	"fmt"
	"strings"

	"github.com/itering/scale.go/utiles"
)

type Result struct {
	ScaleDecoder
}

func (b *Result) Process() {
	optionValue := utiles.BytesToHex(b.NextBytes(1))
	subType := strings.Split(b.SubType, ",")
	if len(subType) != 2 {
		panic("Result subType not illegal")
	}
	if optionValue == "00" || optionValue == "" {
		b.Value = map[string]interface{}{"Ok": b.ProcessAndUpdateData(subType[0])}
	} else if optionValue == "01" {
		b.Value = map[string]interface{}{"Error": b.ProcessAndUpdateData(subType[1])}
	} else {
		panic("illegal Result data")
	}
}

func (b *Result) Encode(value interface{}) string {
	typed, ok := value.(map[string]interface{})
	if !ok {
		panic("invalid Result input")
	}
	subType := strings.Split(b.SubType, ",")
	if len(subType) != 2 {
		panic("Result subType not illegal")
	}
	if data, ok := typed["Ok"]; ok {
		return "00" + EncodeWithOpt(subType[0], data, &ScaleDecoderOption{Spec: b.Spec, Metadata: b.Metadata})
	}
	if data, ok := typed["Error"]; ok {
		return "01" + EncodeWithOpt(subType[1], data, &ScaleDecoderOption{Spec: b.Spec, Metadata: b.Metadata})
	}
	panic("illegal Result data")
}

func (b *Result) TypeStructString() string {
	subType := strings.Split(b.SubType, ",")
	return fmt.Sprintf("Results<%s,%s>", getTypeStructString(subType[0], 0), getTypeStructString(subType[1], 0))
}

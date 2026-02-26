package types

import (
	"fmt"
	"strings"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type Vec struct {
	ScaleDecoder
}

const maxVecLength = 512 * 1024

func (v *Vec) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	if v.SubType != "" && option != nil {
		option.SubType = v.SubType
	}
	v.ScaleDecoder.Init(data, option)
}

func (v *Vec) Process() {
	elementCount := v.ProcessAndUpdateData("Compact<u32>").(int)
	var result []interface{}
	if elementCount > maxVecLength {
		panic(fmt.Sprintf("Vec length %d exceeds %d with subType %s", elementCount, maxVecLength, v.SubType))
	}
	subType := v.SubType
	for i := 0; i < elementCount; i++ {
		result = append(result, v.ProcessAndUpdateData(subType))
	}
	v.Value = result
}

func (v *Vec) Encode(value interface{}) string {
	var raw string
	if v, ok := value.(string); ok {
		if v == "" {
			return Encode("Compact<u32>", 0)
		}
		panic(fmt.Errorf("invalid vec input"))
	}
	if value == nil {
		return Encode("Compact<u32>", 0)
	}
	values, ok := asInterfaceSlice(value)
	if !ok {
		panic(fmt.Errorf("invalid vec input"))
	}
	raw += Encode("Compact<u32>", len(values))
	for _, item := range values {
		raw += utiles.TrimHex(EncodeWithOpt(v.SubType, item, &ScaleDecoderOption{Spec: v.Spec, Metadata: v.Metadata}))
	}
	return raw
}

func (v *Vec) TypeStructString() string {
	return fmt.Sprintf("Vec<%s>", getTypeStructString(v.SubType, v.RecursiveTime))
}

type BoundedVec struct {
	Vec
}

// Init BoundedVec<Type, Size> to Vec<Type>
// https://github.com/paritytech/substrate/pull/8556
func (v *BoundedVec) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	if option != nil {
		if BoundedArr := strings.Split(option.SubType, ","); len(BoundedArr) >= 2 {
			size := BoundedArr[len(BoundedArr)-1]
			v.SubType = strings.Replace(option.SubType, fmt.Sprintf(",%s", size), "", 1)
			option.SubType = v.SubType
		}
	}
	v.ScaleDecoder.Init(data, option)
}

// WeakBoundedVec https://github.com/paritytech/substrate/pull/8842
type WeakBoundedVec struct {
	BoundedVec
}

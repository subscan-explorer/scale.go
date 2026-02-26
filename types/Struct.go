package types

import (
	"fmt"
	"strings"
)

type Struct struct {
	ScaleDecoder
}

func (s *Struct) WithTypeMapping(typeMapping *TypeMapping) *Struct {
	s.TypeMapping = typeMapping
	return s
}

func (s *Struct) Process() {
	result := make(map[string]interface{})
	if s.TypeMapping != nil {
		for k, v := range s.TypeMapping.Names {
			result[v] = s.ProcessAndUpdateData(s.TypeMapping.Types[k])
		}
	}
	s.Value = result
}

func (s *Struct) Encode(value interface{}) string {
	var raw string
	typed, ok := value.(map[string]interface{})
	if !ok {
		panic("invalid struct input")
	}
	if s.TypeMapping != nil {
		for k, v := range s.TypeMapping.Names {
			raw += EncodeWithOpt(s.TypeMapping.Types[k], typed[v], &ScaleDecoderOption{Spec: s.Spec, Metadata: s.Metadata})
		}
	}
	return raw
}

func (s *Struct) TypeStructString() string {
	if s.TypeMapping != nil {
		var typeStrings []string
		s.RecursiveTime++
		for _, subType := range s.TypeMapping.Types {
			if s.RecursiveTime > limitRecursiveTime {
				return "Struct(...)"
			}
			typeStrings = append(typeStrings, getTypeStructString(subType, s.RecursiveTime))
		}
		return fmt.Sprintf("Struct(%s)", strings.Join(typeStrings, ","))
	}
	return s.ScaleDecoder.TypeString
}

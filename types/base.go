package types

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

const limitRecursiveTime = 5

type ScaleDecoderOption struct {
	Spec             int
	SubType          string
	Module           string
	ValueList        []string
	Metadata         *MetadataStruct
	FixedLength      int
	SignedExtensions []SignedExtension `json:"signed_extensions"`
	AdditionalCheck  []string
	TypeName         string
	recursiveTime    int
}

type TypeMapping struct {
	Names []string
	Types []string
}

type SignedExtension struct {
	Name             string             `json:"name"`
	AdditionalSigned []AdditionalSigned `json:"additional_signed"`
}

type AdditionalSigned struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Decoder interface {
	Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption)
	Process()
	TypeStructString() string
	GetData() scaleBytes.ScaleBytes
	GetInternalCall() []string
	GetValue() interface{}
}

type TypeMappingGetter interface {
	GetTypeMapping() *TypeMapping
}

type Encoder interface {
	Encode(interface{}) string
}

type ScaleDecoder struct {
	Data             scaleBytes.ScaleBytes `json:"-"`
	TypeString       string                `json:"-"`
	SubType          string                `json:"-"`
	Value            interface{}           `json:"-"`
	RawValue         string                `json:"-"`
	TypeMapping      *TypeMapping          `json:"-"`
	Metadata         *MetadataStruct       `json:"-"`
	Spec             int                   `json:"-"`
	Module           string                `json:"-"`
	DuplicateName    map[string]int        `json:"-"`
	TypeName         string                `json:"-"`
	RegisteredSiType map[int]string        `json:"-"`
	InternalCall     []string              `json:"-"`
	RecursiveTime    int                   `json:"-"`
}

func (s *ScaleDecoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	if option != nil {
		if option.Metadata != nil {
			s.Metadata = option.Metadata
		}
		if option.SubType != "" {
			s.SubType = option.SubType
		}
		if option.Spec != 0 {
			s.Spec = option.Spec
		}
		if option.Module != "" {
			s.Module = option.Module
		}
		if option.TypeName != "" {
			s.TypeName = option.TypeName
		}
		s.RecursiveTime = option.recursiveTime
	}
	if len(s.DuplicateName) == 0 {
		s.DuplicateName = make(map[string]int)
	}
	s.Data = data
	s.RawValue = ""
	s.Value = nil
	if s.TypeMapping == nil && s.TypeString != "" {
		s.buildStruct()
	}
}

func (s *ScaleDecoder) Process() {}

func (s *ScaleDecoder) Encode(interface{}) string {
	panic(fmt.Sprintf("not found base type %s", s.TypeName))
}

func (s *ScaleDecoder) GetData() scaleBytes.ScaleBytes {
	return s.Data
}

func (s *ScaleDecoder) GetInternalCall() []string {
	return s.InternalCall
}

func (s *ScaleDecoder) GetValue() interface{} {
	return s.Value
}

func (s *ScaleDecoder) GetTypeMapping() *TypeMapping {
	return s.TypeMapping
}

// TypeStructString Type Struct string
func (s *ScaleDecoder) TypeStructString() string {
	return s.TypeName
}

func (s *ScaleDecoder) NextBytes(length int) []byte {
	data := s.Data.GetNextBytes(length)
	s.RawValue += utiles.BytesToHex(data)
	return data
}

func (s *ScaleDecoder) GetNextU8() int {
	b := s.NextBytes(1)
	if len(b) > 0 {
		return int(b[0])
	}
	return 0
}

func (s *ScaleDecoder) getNextBool() bool {
	data := s.NextBytes(1)
	return utiles.BytesToHex(data) == "01"
}

// func (s *ScaleDecoder) reset() {
// 	s.Data.Data = []byte{}
// 	s.Data.Offset = 0
// }

func (s *ScaleDecoder) buildStruct() {
	if s.TypeString != "" && string(s.TypeString[0]) == "(" && s.TypeString[len(s.TypeString)-1:] == ")" {

		var names, types []string
		reg := regexp.MustCompile(`[\<\(](.*?)[\>\)]`)
		typeString := s.TypeString[1 : len(s.TypeString)-1]
		typeParts := reg.FindAllString(typeString, -1)
		for _, part := range typeParts {
			typeString = strings.ReplaceAll(typeString, part, strings.ReplaceAll(part, ",", "#"))
		}

		for k, v := range strings.Split(typeString, ",") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			types = append(types, strings.ReplaceAll(strings.TrimSpace(v), "#", ","))
			names = append(names, fmt.Sprintf("col%d", k+1))
		}

		s.TypeMapping = &TypeMapping{Names: names, Types: types}
	}
}

func (s *ScaleDecoder) ProcessAndUpdateData(typeString string) interface{} {
	r := RuntimeType{Module: s.Module}
	if value, ok := s.fastProcess(typeString); ok {
		return value
	}

	decoder, subType, err := r.GetCodec(typeString, s.Spec)
	if err != nil {
		panic(fmt.Sprintf("Not found decoder class %s", typeString))
	}

	offsetStart := s.Data.Offset

	// init
	option := ScaleDecoderOption{SubType: subType, Spec: s.Spec, Metadata: s.Metadata, Module: s.Module, TypeName: typeString}
	decoder.Init(s.Data, &option)

	// process do decode
	decoder.Process()
	elementData := decoder.GetData()
	if internalCall := decoder.GetInternalCall(); len(internalCall) > 0 {
		s.InternalCall = append(s.InternalCall, internalCall...)
	}
	s.Data.Offset = elementData.Offset
	s.Data.Data = elementData.Data
	s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])

	return decoder.GetValue()
}

func (s *ScaleDecoder) fastProcess(typeString string) (interface{}, bool) {
	switch strings.ToLower(typeString) {
	case "u8":
		offsetStart := s.Data.Offset
		data := s.Data.GetNextBytes(1)
		var value int
		if len(data) > 0 {
			value = int(data[0])
		}
		s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])
		return value, true
	case "u16":
		offsetStart := s.Data.Offset
		data := s.Data.GetNextBytes(2)
		var c [2]byte
		copy(c[:], data)
		s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])
		return binary.LittleEndian.Uint16(c[:]), true
	case "u32":
		offsetStart := s.Data.Offset
		data := s.Data.GetNextBytes(4)
		var c [4]byte
		copy(c[:], data)
		s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])
		return binary.LittleEndian.Uint32(c[:]), true
	case "u64":
		offsetStart := s.Data.Offset
		data := s.Data.GetNextBytes(8)
		var c [8]byte
		copy(c[:], data)
		s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])
		return binary.LittleEndian.Uint64(c[:]), true
	case "bool":
		offsetStart := s.Data.Offset
		data := s.Data.GetNextBytes(1)
		value := len(data) > 0 && data[0] == 1
		s.RawValue = utiles.BytesToHex(s.Data.Data[offsetStart:s.Data.Offset])
		return value, true
	default:
		return nil, false
	}
}

func Encode(typeString string, data interface{}) string {
	return EncodeWithOpt(typeString, data, nil)
}

func EncodeWithOpt(typeString string, data interface{}, opt *ScaleDecoderOption) string {
	r := RuntimeType{}
	if strings.EqualFold(typeString, "Null") {
		return ""
	}
	if opt == nil {
		opt = &ScaleDecoderOption{Spec: -1}
	}
	opt.TypeName = typeString
	decoder, subType, err := r.GetCodec(typeString, opt.Spec)
	if err != nil {
		panic(fmt.Sprintf("Not found decoder class %s", typeString))
	}
	opt.SubType = subType
	decoder.Init(scaleBytes.EmptyScaleBytes(), opt)
	dataVal := data
	if dataVal == nil {
		dataVal = ""
	}
	encoder, ok := decoder.(Encoder)
	if !ok {
		panic(fmt.Sprintf("%s not implement Encode function", typeString))
	}
	return utiles.TrimHex(strings.ToLower(encoder.Encode(dataVal)))
}

func EqTypeStringWithTypeStruct(typeString string, dest *source.TypeStruct) bool {
	typeName := getTypeStructString(typeString, 0)
	if typeName == "" {
		return true
	}
	switch dest.Type {
	case "struct":
		var typeStrings []string
		for _, v := range dest.TypeMapping {
			typeStrings = append(typeStrings, v[1])
		}
		return typeName == strings.Join(typeStrings, "")
	case "enum":
		if len(dest.ValueList) > 0 {
			return typeName == strings.Join(dest.ValueList, "")
		}
		var typeStrings []string
		for _, v := range dest.TypeMapping {
			typeStrings = append(typeStrings, v[1])
		}
		return typeName == strings.Join(typeStrings, "")
	case "string":
		return typeName == getTypeStructString(dest.TypeString, 0)
	}
	return true
}

// getTypeStructString get type struct string
func getTypeStructString(typeString string, recursiveTime int) string {
	r := RuntimeType{}
	decoder, subType, err := r.GetCodec(typeString, 0)
	if err != nil {
		return ""
	}
	opt := &ScaleDecoderOption{SubType: subType, TypeName: typeString, recursiveTime: recursiveTime}
	decoder.Init(scaleBytes.EmptyScaleBytes(), opt)
	return decoder.TypeStructString()
}

// Eq check type string is equal
func Eq(typeString, destTypeString string) bool {
	return strings.EqualFold(getTypeStructString(typeString, 0), getTypeStructString(destTypeString, 0))
}

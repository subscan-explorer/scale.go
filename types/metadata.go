package types

import (
	"encoding/json"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type MetadataModules struct {
	Name                  string                         `json:"name"`
	Prefix                string                         `json:"prefix"`
	Storage               []MetadataStorage              `json:"storage"`
	Calls                 []MetadataCalls                `json:"calls,omitempty"`
	CallsValue            *PalletLookUp                  `json:"calls_value,omitempty"`
	CallsDeprecationInfo  map[int]DeprecationInfo        `json:"calls_deprecation_info,omitempty"`
	Events                []MetadataEvents               `json:"events,omitempty"`
	EventsValue           *PalletLookUp                  `json:"events_value,omitempty"`
	EventsDeprecationInfo map[int]DeprecationInfo        `json:"events_deprecation_info,omitempty"`
	Constants             []MetadataConstants            `json:"constants,omitempty"`
	Errors                []MetadataModuleError          `json:"errors"`
	ErrorsValue           *PalletLookUp                  `json:"errors_value"`
	ErrorsDeprecationInfo map[int]DeprecationInfo        `json:"errors_deprecation_info,omitempty"`
	AssociatedTypes       []PalletAssociatedTypeMetadata `json:"associated_types,omitempty"`
	ViewFunctions         []PalletViewFunctionMetadata   `json:"view_functions,omitempty"`
	DeprecationInfo       *DeprecationInfo               `json:"deprecation_info,omitempty"`
	Index                 int                            `json:"index"`
	Docs                  []string                       `json:"docs,omitempty"`
}

type MetadataStorage struct {
	Name            string           `json:"name"`
	Modifier        string           `json:"modifier"`
	Type            StorageType      `json:"type"`
	Fallback        string           `json:"fallback"`
	Docs            []string         `json:"docs"`
	Hasher          string           `json:"hasher,omitempty"`
	DeprecationInfo *DeprecationInfo `json:"deprecation_info,omitempty"`
}

type StorageType struct {
	Origin         string    `json:"origin"`
	PlainType      *string   `json:"plain_type,omitempty"`
	PlainTypeValue *int      `json:"PlainTypeValue,omitempty"`
	MapType        *MapType  `json:"map_type,omitempty"`
	DoubleMapType  *MapType  `json:"double_map_type,omitempty"`
	NMapType       *NMapType `json:"n_map_type,omitempty"`
}

func (s *StorageType) TypeValue() string {
	switch s.Origin {
	case "MapType":
		return s.MapType.Value
	case "DoubleMapType":
		return s.DoubleMapType.Value
	case "Map":
		return s.NMapType.Value
	default:
		if s.PlainType != nil {
			return *s.PlainType
		}
		return ""
	}
}

type MapType struct {
	Key        string `json:"key"`
	Key2       string `json:"key2,omitempty"`
	Hasher     string `json:"hasher"`
	Key2Hasher string `json:"key2Hasher,omitempty"`
	Value      string `json:"value"`
	IsLinked   bool   `json:"isLinked"`
}

type MapTypeValue struct {
	Key        int    `json:"key"`
	Key2       int    `json:"key2,omitempty"`
	Hasher     string `json:"hasher"`
	Key2Hasher string `json:"key2Hasher,omitempty"`
	Value      int    `json:"value"`
}

type NMapType struct {
	Hashers []string `json:"hashers"`
	KeyVec  []string `json:"key_vec"`
	Value   string   `json:"value"`
	KeysId  int      `json:"keys_id"`
	ValueId int      `json:"value_id"`
}

type MetadataCalls struct {
	Lookup          string                       `json:"lookup"`
	Name            string                       `json:"name"`
	Docs            []string                     `json:"docs"`
	Args            []MetadataModuleCallArgument `json:"args"`
	DeprecationInfo *DeprecationInfo             `json:"deprecation_info,omitempty"`
	LookupIndex     int                          `json:"-"`
}

type MetadataEvents struct {
	Lookup          string           `json:"lookup"`
	Name            string           `json:"name"`
	Docs            []string         `json:"docs"`
	Args            []string         `json:"args"`
	ArgsName        []string         `json:"args_name,omitempty"`
	ArgsTypeName    []string         `json:"args_type_name,omitempty"`
	DeprecationInfo *DeprecationInfo `json:"deprecation_info,omitempty"`
	LookupIndex     int              `json:"-"`
}

type MetadataStruct struct {
	MetadataVersion int                   `json:"metadata_version"`
	Metadata        MetadataTag           `json:"metadata"`
	CallIndex       map[string]CallIndex  `json:"call_index"`
	EventIndex      map[string]EventIndex `json:"event_index"`
	Extrinsic       *ExtrinsicMetadata    `json:"extrinsic"`
	Type            *int                  `json:"type,omitempty"`
	Apis            []RuntimeApiMetadata  `json:"apis,omitempty"`
	OuterEnums      *OuterEnumsMetadata   `json:"outer_enums,omitempty"`
	Customer        *CustomMetadata       `json:"customer,omitempty"`
}

type RuntimeApiMetadata struct {
	Name            string                     `json:"name"`
	Methods         []RuntimeApiMethodMetadata `json:"methods"`
	Docs            []string                   `json:"docs"`
	Version         int                        `json:"version,omitempty"`
	DeprecationInfo *DeprecationInfo           `json:"deprecation_info,omitempty"`
}

type RuntimeApiMethodMetadata struct {
	Name            string                          `json:"name"`
	Inputs          []RuntimeApiMethodParamMetadata `json:"inputs"`
	OutputsId       int                             `json:"outputsId"`
	Outputs         string                          `json:"outputs"`
	Docs            []string                        `json:"docs"`
	DeprecationInfo *DeprecationInfo                `json:"deprecation_info,omitempty"`
}

type RuntimeApiMethodParamMetadata struct {
	Name   string `json:"name"`
	TypeId int    `json:"typeId"`
	Type   string `json:"type"`
}

type ExtrinsicMetadata struct {
	Type                           int                              `json:"type,omitempty"`
	Version                        int                              `json:"version"`
	Versions                       string                           `json:"versions,omitempty"`
	AddressType                    int                              `json:"addressType,omitempty"`
	CallType                       int                              `json:"callType,omitempty"`
	SignatureType                  int                              `json:"signatureType,omitempty"`
	ExtraType                      int                              `json:"extraType,omitempty"`
	SignedExtensions               []SignedExtensions               `json:"signedExtensions"`
	SignedIdentifier               []string                         `json:"signed_identifier"`
	TransactionExtensionsByVersion []TransactionExtensionsByVersion `json:"transaction_extensions_by_version,omitempty"`
	TransactionExtensions          []TransactionExtensionMetadata   `json:"transaction_extensions,omitempty"`
}

type DeprecationInfo struct {
	Type  string  `json:"type"`
	Note  string  `json:"note,omitempty"`
	Since *string `json:"since,omitempty"`
}

type PalletAssociatedTypeMetadata struct {
	Name   string   `json:"name"`
	TypeId int      `json:"type_id"`
	Type   string   `json:"type,omitempty"`
	Docs   []string `json:"docs"`
}

type PalletViewFunctionMetadata struct {
	ID              string                          `json:"id"`
	Name            string                          `json:"name"`
	Inputs          []RuntimeApiMethodParamMetadata `json:"inputs"`
	OutputId        int                             `json:"output_id"`
	Output          string                          `json:"output,omitempty"`
	Docs            []string                        `json:"docs"`
	DeprecationInfo *DeprecationInfo                `json:"deprecation_info,omitempty"`
}

type TransactionExtensionsByVersion struct {
	Version    int   `json:"version"`
	Extensions []int `json:"extensions"`
}

type TransactionExtensionMetadata struct {
	Identifier     string `json:"identifier"`
	Type           int    `json:"type"`
	TypeString     string `json:"type_string,omitempty"`
	Implicit       int    `json:"implicit"`
	ImplicitString string `json:"implicit_string,omitempty"`
}

type OuterEnumsMetadata struct {
	CallType  int `json:"callType"`
	EventType int `json:"eventType"`
	ErrorType int `json:"errorType"`
}

type CustomMetadata struct {
	Map []map[string]CustomValueMetadata `json:"map"`
}

type CustomValueMetadata struct {
	Value string `json:"value"`
	Type  int    `json:"type"`
}

type CallIndex struct {
	Module MetadataModules `json:"module"`
	Call   MetadataCalls   `json:"call"`
}

type EventIndex struct {
	Module MetadataModules `json:"module"`
	Call   MetadataEvents  `json:"call"`
}

type MetadataTag struct {
	Modules []MetadataModules `json:"modules"`
}

type MetadataConstants struct {
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	TypeValue       int              `json:"type_value"`
	ConstantsValue  string           `json:"constants_value"`
	Docs            []string         `json:"docs"`
	DeprecationInfo *DeprecationInfo `json:"deprecation_info,omitempty"`
}

type ExtrinsicParam struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type MetadataModuleCall struct {
	ScaleDecoder
	Name string                       `json:"name"`
	Args []MetadataModuleCallArgument `json:"args"`
	Docs []string                     `json:"docs"`
}

func (m *MetadataModuleCall) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	argsValue := m.ProcessAndUpdateData("Vec<MetadataModuleCallArgument>").([]interface{})
	var args []MetadataModuleCallArgument
	for _, v := range argsValue {
		arg := v.(map[string]string)
		args = append(args, MetadataModuleCallArgument{Name: arg["name"], Type: arg["type"]})
	}
	m.Args = args
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		m.Docs = append(m.Docs, v.(string))
	}
	m.Value = MetadataModuleCall{
		Name: m.Name,
		Args: m.Args,
		Docs: m.Docs,
	}
}

type MetadataModuleCallArgument struct {
	ScaleDecoder
	Name     string `json:"name"`
	Type     string `json:"type"`
	TypeName string `json:"type_name,omitempty"`
}

func (m *MetadataModuleCallArgument) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	m.Type = convert.ConvertType(m.ProcessAndUpdateData("String").(string))
	m.Value = map[string]string{
		"name": m.Name,
		"type": m.Type,
	}
}

type MetadataModuleEvent struct {
	ScaleDecoder
	Name string   `json:"name"`
	Docs []string `json:"docs"`
	Args []string `json:"args"`
}

func (m *MetadataModuleEvent) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	args := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range args {
		m.Args = append(m.Args, v.(string))
	}
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		m.Docs = append(m.Docs, v.(string))
	}
	m.Value = MetadataEvents{
		Name: m.Name,
		Args: m.Args,
		Docs: m.Docs,
	}
}

type MetadataV6Decoder struct {
	ScaleDecoder
}

func (m *MetadataV6Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV6Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	metadataV6ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV6Module>").([]interface{})

	callModuleIndex := 0
	eventModuleIndex := 0
	bm, _ := json.Marshal(metadataV6ModuleCall)
	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{
					Module: module,
					Call:   call,
				}
			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{
					Module: module,
					Call:   event,
				}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV6Module struct {
	ScaleDecoder
	Name       string                   `json:"name"`
	Prefix     string                   `json:"prefix"`
	CallIndex  string                   `json:"call_index"`
	HasStorage bool                     `json:"has_storage"`
	Storage    []MetadataStorage        `json:"storage"`
	HasCalls   bool                     `json:"has_calls"`
	Calls      []MetadataModuleCall     `json:"calls"`
	HasEvents  bool                     `json:"has_events"`
	Events     []MetadataEvents         `json:"events"`
	Constants  []map[string]interface{} `json:"constants"`
}

func (m *MetadataV6Module) GetIdentifier() string {
	return m.Name
}

func (m *MetadataV6Module) Process() {
	cm := MetadataV6Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)
	cm.Prefix = m.ProcessAndUpdateData("String").(string)
	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("Vec<MetadataV6ModuleStorage>").([]interface{})
		var storage []MetadataStorage
		for _, v := range storageValue {
			storage = append(storage, v.(MetadataStorage))
		}
		cm.Storage = storage
	}

	cm.HasCalls = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasCalls {
		callValue := m.ProcessAndUpdateData("Vec<MetadataModuleCall>").([]interface{})
		var calls []MetadataModuleCall
		for _, v := range callValue {
			calls = append(calls, v.(MetadataModuleCall))
		}
		cm.Calls = calls
	}
	cm.HasEvents = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasEvents {
		eventValue := m.ProcessAndUpdateData("Vec<MetadataModuleEvent>").([]interface{})
		var events []MetadataEvents
		for _, v := range eventValue {
			events = append(events, v.(MetadataEvents))
		}
		cm.Events = events
	}
	constantValue := m.ProcessAndUpdateData("Vec<MetadataV6ModuleConstants>").([]interface{})
	var constants []map[string]interface{}
	for _, v := range constantValue {
		constants = append(constants, v.(map[string]interface{}))
	}
	cm.Constants = constants
	m.Value = cm
}

type MetadataV6ModuleConstants struct {
	ScaleDecoder
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ConstantsValue string   `json:"constants_value"`
	Docs           []string `json:"docs"`
}

func (m *MetadataV6ModuleConstants) Process() {
	name := m.ProcessAndUpdateData("String").(string)
	cType := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
	ConstantsValue := m.ProcessAndUpdateData("HexBytes").(string)
	var docsArr []string
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		docsArr = append(docsArr, v.(string))
	}
	r := map[string]interface{}{
		"name":            name,
		"type":            cType,
		"constants_value": ConstantsValue,
		"docs":            docsArr,
	}
	m.Value = r
}

type MetadataV6ModuleStorage struct {
	ScaleDecoder
	Name     string                 `json:"name"`
	Modifier string                 `json:"modifier"`
	Type     map[string]interface{} `json:"type"`
	Fallback string                 `json:"fallback"`
	Docs     []string               `json:"docs"`
	Hasher   string                 `json:"hasher"`
}

func (m *MetadataV6ModuleStorage) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV6ModuleStorage) Process() {
	cm := MetadataStorage{}
	cm.Name = m.ProcessAndUpdateData("String").(string)
	cm.Modifier = m.ProcessAndUpdateData("StorageModify").(string)
	storageFunctionType := m.ProcessAndUpdateData("StorageFunctionType").(string)
	if storageFunctionType == "MapType" {
		cm.Hasher = m.ProcessAndUpdateData("StorageHasher").(string)
		cm.Type = StorageType{
			Origin: "MapType",
			MapType: &MapType{
				Hasher:   cm.Hasher,
				Key:      convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				Value:    convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				IsLinked: m.ProcessAndUpdateData("bool").(bool),
			},
		}
	} else if storageFunctionType == "DoubleMapType" {
		cm.Hasher = m.ProcessAndUpdateData("StorageHasher").(string)
		key1 := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		key2 := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		value := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		key2Hasher := m.ProcessAndUpdateData("StorageHasher").(string)
		cm.Type = StorageType{
			Origin: "DoubleMapType",
			DoubleMapType: &MapType{
				Hasher:     cm.Hasher,
				Key:        key1,
				Key2:       key2,
				Value:      value,
				Key2Hasher: key2Hasher,
			},
		}
	} else if storageFunctionType == "PlainType" {
		plainType := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		cm.Type = StorageType{Origin: "PlainType", PlainType: &plainType}
	}
	cm.Fallback = m.ProcessAndUpdateData("HexBytes").(string)
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		cm.Docs = append(m.Docs, v.(string))
	}
	m.Value = cm
}

type MetadataV7Decoder struct {
	ScaleDecoder
}

func (m *MetadataV7Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV7Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	metadataV7ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV7Module>").([]interface{})
	callModuleIndex := 0
	eventModuleIndex := 0
	bm, _ := json.Marshal(metadataV7ModuleCall)
	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{
					Module: module,
					Call:   call,
				}
			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{
					Module: module,
					Call:   event,
				}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV7Module struct {
	ScaleDecoder
	Name       string                   `json:"name"`
	Prefix     string                   `json:"prefix"`
	CallIndex  string                   `json:"call_index"`
	HasStorage bool                     `json:"has_storage"`
	Storage    []MetadataStorage        `json:"storage"`
	HasCalls   bool                     `json:"has_calls"`
	Calls      []MetadataModuleCall     `json:"calls"`
	HasEvents  bool                     `json:"has_events"`
	Events     []MetadataEvents         `json:"events"`
	Constants  []map[string]interface{} `json:"constants"`
}

func (m *MetadataV7Module) Process() {
	cm := MetadataV7Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)

	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("MetadataV7ModuleStorage").(MetadataV7ModuleStorage)
		cm.Storage = storageValue.Items
		cm.Prefix = storageValue.Prefix
	}
	cm.HasCalls = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasCalls {
		callValue := m.ProcessAndUpdateData("Vec<MetadataModuleCall>").([]interface{})
		var calls []MetadataModuleCall
		for _, v := range callValue {
			calls = append(calls, v.(MetadataModuleCall))
		}
		cm.Calls = calls
	}
	cm.HasEvents = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasEvents {
		eventValue := m.ProcessAndUpdateData("Vec<MetadataModuleEvent>").([]interface{})
		var events []MetadataEvents
		for _, v := range eventValue {
			events = append(events, v.(MetadataEvents))
		}
		cm.Events = events
	}
	constantValue := m.ProcessAndUpdateData("Vec<MetadataV7ModuleConstants>").([]interface{})
	var constants []map[string]interface{}
	for _, v := range constantValue {
		constants = append(constants, v.(map[string]interface{}))
	}
	cm.Constants = constants
	m.Value = cm
}

type MetadataV7ModuleConstants struct {
	MetadataV6ModuleConstants
}

type MetadataV7ModuleStorage struct {
	MetadataV6ModuleStorage
	Prefix string            `json:"prefix"`
	Items  []MetadataStorage `json:"items"`
}

func (m *MetadataV7ModuleStorage) Process() {
	cm := MetadataV7ModuleStorage{}
	cm.Prefix = m.ProcessAndUpdateData("String").(string)
	itemsValue := m.ProcessAndUpdateData("Vec<MetadataV7ModuleStorageEntry>").([]interface{})
	var items []MetadataStorage
	for _, v := range itemsValue {
		items = append(items, v.(MetadataStorage))
	}
	cm.Items = items
	m.Value = cm
}

type MetadataV7ModuleStorageEntry struct {
	ScaleDecoder
	Name     string      `json:"name"`
	Modifier string      `json:"modifier"`
	Type     StorageType `json:"type"`
	Fallback string      `json:"fallback"`
	Docs     []string    `json:"docs"`
	Hasher   string      `json:"hasher"`
}

func (m *MetadataV7ModuleStorageEntry) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV7ModuleStorageEntry) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	m.Modifier = m.ProcessAndUpdateData("StorageModify").(string)
	storageFunctionType := m.ProcessAndUpdateData("StorageFunctionType").(string)
	if storageFunctionType == "MapType" {
		m.Hasher = m.ProcessAndUpdateData("StorageHasher").(string)
		m.Type = StorageType{
			Origin: "MapType",
			MapType: &MapType{
				Hasher:   m.Hasher,
				Key:      convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				Value:    convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				IsLinked: m.ProcessAndUpdateData("bool").(bool)},
		}
	} else if storageFunctionType == "DoubleMapType" {
		m.Hasher = m.ProcessAndUpdateData("StorageHasher").(string)
		key1 := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		key2 := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		value := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		key2Hasher := m.ProcessAndUpdateData("StorageHasher").(string)
		m.Type = StorageType{
			Origin: "DoubleMapType",
			DoubleMapType: &MapType{
				Hasher:     m.Hasher,
				Key:        key1,
				Key2:       key2,
				Value:      value,
				Key2Hasher: key2Hasher,
			},
		}
	} else if storageFunctionType == "PlainType" {
		plainType := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		m.Type = StorageType{
			Origin:    "PlainType",
			PlainType: &plainType}
	}
	m.Fallback = m.ProcessAndUpdateData("HexBytes").(string)
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		m.Docs = append(m.Docs, v.(string))
	}
	m.Value = MetadataStorage{
		Name:     m.Name,
		Modifier: m.Modifier,
		Type:     m.Type,
		Fallback: m.Fallback,
		Docs:     m.Docs,
	}
}

type MetadataV8Decoder struct {
	ScaleDecoder
}

func (m *MetadataV8Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV8Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	metadataV8ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV8Module>").([]interface{})
	callModuleIndex := 0
	eventModuleIndex := 0
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	bm, _ := json.Marshal(metadataV8ModuleCall)
	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{
					Module: module,
					Call:   call,
				}
			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{
					Module: module,
					Call:   event,
				}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV8Module struct {
	ScaleDecoder
	Name       string                   `json:"name"`
	Prefix     string                   `json:"prefix"`
	CallIndex  string                   `json:"call_index"`
	HasStorage bool                     `json:"has_storage"`
	Storage    []MetadataStorage        `json:"storage"`
	HasCalls   bool                     `json:"has_calls"`
	Calls      []MetadataModuleCall     `json:"calls"`
	HasEvents  bool                     `json:"has_events"`
	Events     []MetadataEvents         `json:"events"`
	Constants  []map[string]interface{} `json:"constants"`
	Errors     []MetadataModuleError    `json:"errors"`
}

func (m *MetadataV8Module) GetIdentifier() string {
	return m.Name
}

func (m *MetadataV8Module) Process() {
	cm := MetadataV8Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)

	// storage
	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("MetadataV7ModuleStorage").(MetadataV7ModuleStorage)
		cm.Storage = storageValue.Items
		cm.Prefix = storageValue.Prefix
	}

	// call
	cm.HasCalls = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasCalls {
		callValue := m.ProcessAndUpdateData("Vec<MetadataModuleCall>").([]interface{})
		calls := []MetadataModuleCall{}
		for _, v := range callValue {
			calls = append(calls, v.(MetadataModuleCall))
		}
		cm.Calls = calls
	}

	// event
	cm.HasEvents = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasEvents {
		eventValue := m.ProcessAndUpdateData("Vec<MetadataModuleEvent>").([]interface{})
		events := []MetadataEvents{}
		for _, v := range eventValue {
			events = append(events, v.(MetadataEvents))
		}
		cm.Events = events
	}

	// constant
	constantValue := m.ProcessAndUpdateData("Vec<MetadataV7ModuleConstants>").([]interface{})
	var constants []map[string]interface{}
	for _, v := range constantValue {
		constants = append(constants, v.(map[string]interface{}))
	}
	cm.Constants = constants

	errorValue := m.ProcessAndUpdateData("Vec<MetadataModuleError>").([]interface{})
	var errors []MetadataModuleError
	for _, v := range errorValue {
		errors = append(errors, v.(MetadataModuleError))
	}
	cm.Errors = errors
	m.Value = cm
}

type MetadataModuleError struct {
	ScaleDecoder    `json:"-"`
	Name            string             `json:"name"`
	Doc             []string           `json:"doc"`
	Fields          []ModuleErrorField `json:"fields,omitempty"`
	DeprecationInfo *DeprecationInfo   `json:"deprecation_info,omitempty"`
	Index           int                `json:"index"`
}

type ModuleErrorField struct {
	Type     string   `json:"type"`
	TypeName string   `json:"type_name"`
	Doc      []string `json:"doc"`
}

func (m *MetadataModuleError) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.Name = ""
	m.Doc = []string{}
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataModuleError) Process() {
	cm := MetadataModuleError{}
	cm.Name = m.ProcessAndUpdateData("String").(string)
	var docsArr []string
	docs := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	for _, v := range docs {
		docsArr = append(docsArr, v.(string))
	}
	cm.Doc = docsArr
	m.Value = cm
}

type MetadataV9Decoder struct {
	ScaleDecoder
}

func (m *MetadataV9Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV9Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	MetadataV9ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV8Module>").([]interface{})
	callModuleIndex := 0
	eventModuleIndex := 0
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	bm, _ := json.Marshal(MetadataV9ModuleCall)
	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{Module: module, Call: call}
			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{Module: module, Call: event}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV10Decoder struct {
	ScaleDecoder
}

func (m *MetadataV10Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV10Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	MetadataV10ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV8Module>").([]interface{})
	callModuleIndex := 0
	eventModuleIndex := 0
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	bm, _ := json.Marshal(MetadataV10ModuleCall)
	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{Module: MetadataModules{Name: module.Name}, Call: call}

			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{Module: MetadataModules{Name: module.Name}, Call: event}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV11Decoder struct {
	ScaleDecoder
}

func (m *MetadataV11Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV11Decoder) Process() {
	var callModuleIndex, eventModuleIndex int

	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	MetadataV11ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV8Module>").([]interface{})
	bm, _ := json.Marshal(MetadataV11ModuleCall)

	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(callModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{Module: MetadataModules{Name: module.Name}, Call: call}
			}
			callModuleIndex++
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(eventModuleIndex), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{Module: MetadataModules{Name: module.Name}, Call: event}
			}
			eventModuleIndex++
		}
	}

	result.Metadata.Modules = modulesType
	m.Value = result
}

type MetadataV12Decoder struct {
	ScaleDecoder
}

type SignedExtensions struct {
	Identifier       string `json:"identifier"`
	Type             int    `json:"type"`
	TypeString       string `json:"-"`
	AdditionalSigned int    `json:"additionalSigned"`
}

type ExtrinsicMetadataV12 struct {
	SignedExtensions []string `json:"signedExtensions"`
}

func (m *MetadataV12Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV12Decoder) Process() {
	// var callModuleIndex, eventModuleIndex int

	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	MetadataV11ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV12Module>").([]interface{})
	bm, _ := json.Marshal(MetadataV11ModuleCall)

	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)
	for k, module := range modulesType {
		if module.Calls != nil {
			for callIndex, call := range module.Calls {
				modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(module.Index), 2, "0") + xstrings.RightJustify(utiles.IntToHex(callIndex), 2, "0")
				result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{Module: MetadataModules{Name: module.Name}, Call: call}
			}
		}
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(module.Index), 2, "0") + xstrings.RightJustify(utiles.IntToHex(eventIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{Module: MetadataModules{Name: module.Name}, Call: event}
			}
		}
	}

	result.Metadata.Modules = modulesType
	extrinsicMetadata := m.ProcessAndUpdateData("ExtrinsicMetadata").(map[string]interface{})
	var extrinsic ExtrinsicMetadataV12
	bm, _ = json.Marshal(extrinsicMetadata)
	_ = json.Unmarshal(bm, &extrinsic)
	result.Extrinsic = &ExtrinsicMetadata{SignedIdentifier: extrinsic.SignedExtensions}
	m.Value = result
}

type MetadataV12Module struct {
	ScaleDecoder
	Name       string                   `json:"name"`
	Prefix     string                   `json:"prefix"`
	CallIndex  string                   `json:"call_index"`
	HasStorage bool                     `json:"has_storage"`
	Storage    []MetadataStorage        `json:"storage"`
	HasCalls   bool                     `json:"has_calls"`
	Calls      []MetadataModuleCall     `json:"calls"`
	HasEvents  bool                     `json:"has_events"`
	Events     []MetadataEvents         `json:"events"`
	Constants  []map[string]interface{} `json:"constants"`
	Errors     []MetadataModuleError    `json:"errors"`
	Index      int                      `json:"index"`
}

func (m *MetadataV12Module) Process() {
	cm := MetadataV12Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)

	// storage
	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("MetadataV7ModuleStorage").(MetadataV7ModuleStorage)
		cm.Storage = storageValue.Items
		cm.Prefix = storageValue.Prefix
	}

	// call
	cm.HasCalls = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasCalls {
		callValue := m.ProcessAndUpdateData("Vec<MetadataModuleCall>").([]interface{})
		var calls []MetadataModuleCall
		for _, v := range callValue {
			calls = append(calls, v.(MetadataModuleCall))
		}
		cm.Calls = calls
	}

	// event
	cm.HasEvents = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasEvents {
		eventValue := m.ProcessAndUpdateData("Vec<MetadataModuleEvent>").([]interface{})
		var events []MetadataEvents
		for _, v := range eventValue {
			events = append(events, v.(MetadataEvents))
		}
		cm.Events = events
	}

	// constant
	constantValue := m.ProcessAndUpdateData("Vec<MetadataV7ModuleConstants>").([]interface{})
	var constants []map[string]interface{}
	for _, v := range constantValue {
		constants = append(constants, v.(map[string]interface{}))
	}
	cm.Constants = constants

	errorValue := m.ProcessAndUpdateData("Vec<MetadataModuleError>").([]interface{})
	var errors []MetadataModuleError
	for _, v := range errorValue {
		errors = append(errors, v.(MetadataModuleError))
	}
	cm.Errors = errors
	cm.Index = m.ProcessAndUpdateData("U8").(int)
	m.Value = cm
}

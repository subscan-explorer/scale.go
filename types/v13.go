package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
)

type MetadataV13Decoder struct {
	ScaleDecoder
}

func (m *MetadataV13Decoder) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

// https://github.com/polkadot-js/api/blob/ddf6f574f616d28cc0c59354baaf58208a776274/packages/metadata/src/v13/toLatest.ts#L14
var KnownOrigins = map[string]string{
	"Council":            "CollectiveOrigin",
	"System":             "SystemOrigin",
	"TechnicalCommittee": "CollectiveOrigin",
	// Polkadot
	"Xcm":       "XcmOrigin",
	"XcmPallet": "XcmOrigin",
	// Acala
	"Authority":      "AuthorityOrigin",
	"GeneralCouncil": "CollectiveOrigin",
}

type OriginCaller struct {
	Name  string
	Index int
}

func (m *MetadataV13Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}
	MetadataV11ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV13Module>").([]interface{})
	bm, _ := json.Marshal(MetadataV11ModuleCall)

	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)

	var originCallers []OriginCaller
	for k, module := range modulesType {
		originCallers = append(originCallers, OriginCaller{Name: module.Name, Index: module.Index})
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
	bm, _ = json.Marshal(extrinsicMetadata)
	_ = json.Unmarshal(bm, &result.Extrinsic)
	registerOriginCaller(originCallers)
	var extrinsic ExtrinsicMetadataV12
	bm, _ = json.Marshal(extrinsicMetadata)
	_ = json.Unmarshal(bm, &extrinsic)
	result.Extrinsic = &ExtrinsicMetadata{SignedIdentifier: extrinsic.SignedExtensions}
	m.Value = result
}

// https://github.com/polkadot-js/api/blob/ddf6f574f616d28cc0c59354baaf58208a776274/packages/metadata/src/v13/toLatest.ts#L117
func registerOriginCaller(originCallers []OriginCaller) {
	sort.Slice(originCallers[:], func(i, j int) bool { return originCallers[i].Index < originCallers[j].Index })
	e := Enum{}
	e.TypeMapping = &TypeMapping{}
	for _, caller := range originCallers {
		for i := len(e.TypeMapping.Names); i < caller.Index; i++ {
			e.TypeMapping.Names = append(e.TypeMapping.Names, fmt.Sprintf("EMPTY%d", i))
			e.TypeMapping.Types = append(e.TypeMapping.Types, "NULL")
		}
		e.TypeMapping.Names = append(e.TypeMapping.Names, caller.Name)
		if t, ok := KnownOrigins[caller.Name]; ok {
			e.TypeMapping.Types = append(e.TypeMapping.Types, t)
		} else {
			e.TypeMapping.Types = append(e.TypeMapping.Types, "NULL")
		}
	}
	typeMapping := e.TypeMapping
	regCustomKey(strings.ToLower("OriginCaller"), func() Decoder {
		origin := Enum{}
		origin.TypeMapping = typeMapping
		return &origin
	})
}

type MetadataV13Module struct {
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

func (m *MetadataV13Module) Process() {
	cm := MetadataV13Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)

	// storage
	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("MetadataV13ModuleStorage").(MetadataV7ModuleStorage)
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

type MetadataV13ModuleStorage struct {
	MetadataV6ModuleStorage
	Prefix string            `json:"prefix"`
	Items  []MetadataStorage `json:"items"`
}

func (m *MetadataV13ModuleStorage) Process() {
	cm := MetadataV7ModuleStorage{}
	cm.Prefix = m.ProcessAndUpdateData("String").(string)
	itemsValue := m.ProcessAndUpdateData("Vec<MetadataV13ModuleStorageEntry>").([]interface{})
	var items []MetadataStorage
	for _, v := range itemsValue {
		items = append(items, v.(MetadataStorage))
	}
	cm.Items = items
	m.Value = cm
}

type MetadataV13ModuleStorageEntry struct {
	ScaleDecoder
	Name     string      `json:"name"`
	Modifier string      `json:"modifier"`
	Type     StorageType `json:"type"`
	Fallback string      `json:"fallback"`
	Docs     []string    `json:"docs"`
	Hasher   string      `json:"hasher"`
}

func (m *MetadataV13ModuleStorageEntry) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	m.ScaleDecoder.Init(data, option)
}

func (m *MetadataV13ModuleStorageEntry) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	m.Modifier = m.ProcessAndUpdateData("StorageModify").(string)
	storageFunctionType := m.ProcessAndUpdateData("StorageFunctionType").(string)

	switch storageFunctionType {
	case "MapType":
		m.Hasher = m.ProcessAndUpdateData("StorageHasher").(string)
		m.Type = StorageType{
			Origin: "MapType",
			MapType: &MapType{
				Hasher:   m.Hasher,
				Key:      convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				Value:    convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
				IsLinked: m.ProcessAndUpdateData("bool").(bool)},
		}
	case "DoubleMapType":
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
	case "PlainType":
		plainType := convert.ConvertType(m.ProcessAndUpdateData("String").(string))
		m.Type = StorageType{
			Origin:    "PlainType",
			PlainType: &plainType}
	case "NMap":
		var KeyVec []string
		for _, v := range m.ProcessAndUpdateData("Vec<String>").([]interface{}) {
			KeyVec = append(KeyVec, convert.ConvertType(v.(string)))
		}
		var hasherVec []string
		for _, v := range m.ProcessAndUpdateData("Vec<StorageHasher>").([]interface{}) {
			hasherVec = append(hasherVec, convert.ConvertType(v.(string)))
		}
		m.Type = StorageType{
			Origin: "NMapType",
			NMapType: &NMapType{
				Hashers: hasherVec,
				KeyVec:  KeyVec,
				Value:   convert.ConvertType(m.ProcessAndUpdateData("String").(string)),
			}}
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

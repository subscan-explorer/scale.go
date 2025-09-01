package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/utiles"
)

// {
//    "lookup": "PortableRegistry",
//    "pallets": "Vec<PalletMetadataV14>",
//    "extrinsic": "ExtrinsicMetadataV14",
// }

type MetadataV14Decoder struct {
	ScaleDecoder
}

type PalletLookUp struct {
	Type int `json:"type"`
}

func (m *MetadataV14Decoder) Process() {
	result := MetadataStruct{
		Metadata: MetadataTag{
			Modules: nil,
		},
	}

	// custom type lookup
	portable := InitPortableRaw(m.ProcessAndUpdateData("PortableRegistry").([]interface{}))
	// utiles.Debug(portable)

	scaleInfo := ScaleInfo{ScaleDecoder: &m.ScaleDecoder, V14: true}
	scaleInfo.ProcessSiType(portable)
	metadataSiType := m.RegisteredSiType
	MetadataV14ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV14Module>").([]interface{})
	bm, _ := json.Marshal(MetadataV14ModuleCall)

	var modulesType []MetadataModules
	_ = json.Unmarshal(bm, &modulesType)
	result.CallIndex = make(map[string]CallIndex)
	result.EventIndex = make(map[string]EventIndex)

	var originCallers []OriginCaller
	for k, module := range modulesType {
		originCallers = append(originCallers, OriginCaller{Name: module.Name, Index: module.Index})

		// calls look up
		if module.CallsValue != nil {
			variants := portable[module.CallsValue.Type].Def.Variant
			if variants == nil {
				panic(fmt.Sprintf("%d call value not variant", module.CallsValue.Type))
			}

			for _, variant := range variants.Variants {
				call := MetadataCalls{Name: variant.Name, Docs: variant.Docs, LookupIndex: variant.Index}
				for _, field := range variant.Fields {
					call.Args = append(call.Args, MetadataModuleCallArgument{
						Name:     field.Name,
						Type:     metadataSiType[field.Type],
						TypeName: convert.ConvertType(field.TypeName),
					})
				}
				module.Calls = append(module.Calls, call)
			}
		}
		modulesType[k].Calls = module.Calls
		for callIndex, call := range module.Calls {
			modulesType[k].Calls[callIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(module.Index), 2, "0") + xstrings.RightJustify(utiles.IntToHex(call.LookupIndex), 2, "0")
			result.CallIndex[modulesType[k].Calls[callIndex].Lookup] = CallIndex{Module: MetadataModules{Name: module.Name}, Call: call}
		}

		// Events
		if module.EventsValue != nil {
			variants := portable[module.EventsValue.Type].Def.Variant
			if variants == nil {
				panic(fmt.Sprintf("%d event value not variant", module.EventsValue.Type))
			}

			for _, variant := range variants.Variants {
				event := MetadataEvents{Name: variant.Name, Docs: variant.Docs, LookupIndex: variant.Index}
				for _, field := range variant.Fields {
					event.Args = append(event.Args, metadataSiType[field.Type])
					event.ArgsTypeName = append(event.ArgsTypeName, convert.ConvertType(field.TypeName))
					event.ArgsName = append(event.ArgsName, field.Name)
				}
				module.Events = append(module.Events, event)
			}
		}
		modulesType[k].Events = module.Events
		if module.Events != nil {
			for eventIndex, event := range module.Events {
				modulesType[k].Events[eventIndex].Lookup = xstrings.RightJustify(utiles.IntToHex(module.Index), 2, "0") + xstrings.RightJustify(utiles.IntToHex(event.LookupIndex), 2, "0")
				result.EventIndex[modulesType[k].Events[eventIndex].Lookup] = EventIndex{Module: MetadataModules{Name: module.Name}, Call: event}
			}
		}

		// Error
		if module.ErrorsValue != nil {
			variants := portable[module.ErrorsValue.Type].Def.Variant
			if variants == nil {
				panic(fmt.Sprintf("%d error value not variant", module.EventsValue.Type))
			}

			for _, variant := range variants.Variants {
				moduleErr := MetadataModuleError{Name: variant.Name, Doc: variant.Docs}
				moduleErr.Index = variant.Index
				for _, field := range variant.Fields {
					moduleErr.Fields = append(moduleErr.Fields, ModuleErrorField{Doc: field.Docs, TypeName: field.TypeName, Type: metadataSiType[field.Type]})
				}
				module.Errors = append(module.Errors, moduleErr)
			}
		}
		modulesType[k].Errors = module.Errors

		// Constant
		for index, constant := range module.Constants {
			variant := metadataSiType[constant.TypeValue]
			if variant == "" {
				panic(fmt.Sprintf("%d constant value not variant", constant.TypeValue))
			}
			module.Constants[index].Type = variant
		}

		// Storage
		for index, storage := range module.Storage {
			if storage.Type.Origin == "PlainType" {
				variant := metadataSiType[*storage.Type.PlainTypeValue]
				module.Storage[index].Type.PlainType = &variant
			} else {
				if maps := storage.Type.NMapType; maps != nil {
					NMapTypeValue := &NMapType{
						Hashers: maps.Hashers,
						Value:   metadataSiType[maps.ValueId],
						KeysId:  maps.KeysId,
						ValueId: maps.ValueId,
					}
					if t := portable[maps.KeysId].Def.Tuple; t != nil {
						if len(NMapTypeValue.Hashers) < len(*t) {
							NMapTypeValue.KeyVec = []string{metadataSiType[maps.KeysId]}
						} else {
							for _, v := range *t {
								NMapTypeValue.KeyVec = append(NMapTypeValue.KeyVec, metadataSiType[v])
							}
						}
					} else {
						NMapTypeValue.KeyVec = TupleDisassemble(metadataSiType[maps.KeysId])
					}
					module.Storage[index].Type.NMapType = NMapTypeValue
				}
			}
		}
	}
	result.Metadata.Modules = modulesType
	extrinsicMetadata := m.ProcessAndUpdateData("ExtrinsicMetadataV14").(map[string]interface{})
	bm, _ = json.Marshal(extrinsicMetadata)
	_ = json.Unmarshal(bm, &result.Extrinsic)

	for index, extension := range result.Extrinsic.SignedExtensions {
		result.Extrinsic.SignedExtensions[index].TypeString = metadataSiType[extension.Type]
		result.Extrinsic.SignedIdentifier = append(result.Extrinsic.SignedIdentifier, extension.Identifier)
	}

	registerOriginCaller(originCallers)
	m.Value = result
}

type MetadataV14Module struct {
	ScaleDecoder
	Name        string                 `json:"name"`
	Prefix      string                 `json:"prefix"`
	CallIndex   string                 `json:"call_index"`
	HasStorage  bool                   `json:"has_storage"`
	Storage     []MetadataStorage      `json:"storage"`
	HasCalls    bool                   `json:"has_calls"`
	Calls       []MetadataModuleCall   `json:"calls"`
	CallsValue  map[string]interface{} `json:"calls_value"`
	HasEvents   bool                   `json:"has_events"`
	Events      []MetadataEvents       `json:"events"`
	EventsValue map[string]interface{} `json:"events_value"`
	Constants   []MetadataConstants    `json:"constants"`
	Errors      []MetadataModuleError  `json:"errors"`
	ErrorsValue map[string]interface{} `json:"errors_value"`
	Index       int                    `json:"index"`
}

func (m *MetadataV14Module) Process() {
	cm := MetadataV14Module{}
	cm.Name = m.ProcessAndUpdateData("String").(string)

	// storage
	cm.HasStorage = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasStorage {
		storageValue := m.ProcessAndUpdateData("MetadataV14ModuleStorage").(MetadataV14ModuleStorage)
		cm.Storage = storageValue.Items
		cm.Prefix = storageValue.Prefix
	}

	// call
	cm.HasCalls = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasCalls {
		cm.CallsValue = m.ProcessAndUpdateData("PalletCallMetadataV14").(map[string]interface{})
	}

	// event
	cm.HasEvents = m.ProcessAndUpdateData("bool").(bool)
	if cm.HasEvents {
		cm.EventsValue = m.ProcessAndUpdateData("PalletEventMetadataV14").(map[string]interface{})
	}

	// constant
	constantValue := m.ProcessAndUpdateData("Vec<PalletConstantMetadataV14>").([]interface{})
	var constants []MetadataConstants
	for _, v := range constantValue {
		constants = append(constants, v.(MetadataConstants))
	}
	cm.Constants = constants

	hasError := m.ProcessAndUpdateData("bool").(bool)
	if hasError {
		cm.ErrorsValue = m.ProcessAndUpdateData("PalletErrorMetadataV14").(map[string]interface{})
	}
	cm.Index = m.ProcessAndUpdateData("U8").(int)
	m.Value = cm
}

type MetadataV14ModuleStorage struct {
	ScaleDecoder
	Prefix string            `json:"prefix"`
	Items  []MetadataStorage `json:"items"`
}

func (m *MetadataV14ModuleStorage) Process() {
	cm := MetadataV14ModuleStorage{}
	cm.Prefix = m.ProcessAndUpdateData("String").(string)
	itemsValue := m.ProcessAndUpdateData("Vec<MetadataV14ModuleStorageEntry>").([]interface{})
	var items []MetadataStorage
	for _, v := range itemsValue {
		items = append(items, v.(MetadataStorage))
	}
	cm.Items = items
	m.Value = cm
}

type MetadataV14ModuleStorageEntry struct {
	ScaleDecoder
	Name     string      `json:"name"`
	Modifier string      `json:"modifier"`
	Type     StorageType `json:"type"`
	Fallback string      `json:"fallback"`
	Docs     []string    `json:"docs"`
	Hasher   string      `json:"hasher"`
}

func (m *MetadataV14ModuleStorageEntry) Process() {
	m.Name = m.ProcessAndUpdateData("String").(string)
	m.Modifier = m.ProcessAndUpdateData("StorageModify").(string)
	storageFunctionType := m.ProcessAndUpdateData("StorageFunctionTypeV14").(string)

	switch storageFunctionType {
	case "PlainType":
		PlainTypeValue := m.ProcessAndUpdateData("SiLookupTypeId").(int)
		m.Type = StorageType{
			Origin:         "PlainType",
			PlainTypeValue: &PlainTypeValue,
		}
	case "Map":
		var hasherVec []string
		for _, v := range m.ProcessAndUpdateData("Vec<StorageHasher>").([]interface{}) {
			hasherVec = append(hasherVec, v.(string))
		}
		m.Type = StorageType{
			Origin: "Map",
			NMapType: &NMapType{
				Hashers: hasherVec,
				KeysId:  m.ProcessAndUpdateData("SiLookupTypeId").(int),
				ValueId: m.ProcessAndUpdateData("SiLookupTypeId").(int),
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

type PalletConstantMetadataV14 struct {
	ScaleDecoder
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ConstantsValue string   `json:"constants_value"`
	Docs           []string `json:"docs"`
}

func (m *PalletConstantMetadataV14) Process() {
	name := m.ProcessAndUpdateData("String").(string)
	ConstantType := m.ProcessAndUpdateData("SiLookupTypeId").(int)
	value := utiles.TrimHex(m.ProcessAndUpdateData("HexBytes").(string))
	docsRaw := m.ProcessAndUpdateData("Vec<String>").([]interface{})
	var docs []string
	for _, v := range docsRaw {
		docs = append(docs, v.(string))
	}
	m.Value = MetadataConstants{
		Name:           name,
		TypeValue:      ConstantType,
		Docs:           docs,
		ConstantsValue: value,
	}
}

func TupleDisassemble(typeString string) []string {
	if typeString == "" {
		panic("nil typeString")
	}
	if string(typeString[0]) == "(" && typeString[len(typeString)-1:] == ")" {
		var types []string
		reg := regexp.MustCompile(`[\<\(](.*?)[\>\)]`)
		typeString := typeString[1 : len(typeString)-1]
		typeParts := reg.FindAllString(typeString, -1)
		for _, part := range typeParts {
			typeString = strings.ReplaceAll(typeString, part, strings.ReplaceAll(part, ",", "#"))
		}
		for _, v := range strings.Split(typeString, ",") {
			types = append(types, strings.ReplaceAll(strings.TrimSpace(v), "#", ","))
		}
		return types

	}
	return []string{typeString}
}

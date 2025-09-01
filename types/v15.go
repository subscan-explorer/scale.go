package types

import (
	"encoding/json"
	"fmt"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/utiles"
)

type MetadataV15Decoder struct {
	ScaleDecoder
}

// "MetadataV15": {
//    "lookup": "PortableRegistry",
//    "pallets": "Vec<PalletMetadataV15>",
//    "extrinsic": "ExtrinsicMetadataV15",
//    "type": "SiLookupTypeId",
//    "apis": "Vec<RuntimeApiMetadataV15>"
//  }

func (m *MetadataV15Decoder) Process() {
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
	MetadataV15ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV15Module>").([]interface{})
	bm, _ := json.Marshal(MetadataV15ModuleCall)

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
				moduleErr := MetadataModuleError{Name: variant.Name, Doc: variant.Docs, Index: variant.Index}
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
						for _, v := range *t {
							NMapTypeValue.KeyVec = append(NMapTypeValue.KeyVec, metadataSiType[v])
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
	extrinsicMetadata := m.ProcessAndUpdateData("ExtrinsicMetadataV15").(map[string]interface{})
	bm, _ = json.Marshal(extrinsicMetadata)
	_ = json.Unmarshal(bm, &result.Extrinsic)

	for index, extension := range result.Extrinsic.SignedExtensions {
		result.Extrinsic.SignedExtensions[index].TypeString = metadataSiType[extension.Type]
		result.Extrinsic.SignedIdentifier = append(result.Extrinsic.SignedIdentifier, extension.Identifier)
	}

	registerOriginCaller(originCallers)

	// type
	metadataType := m.ProcessAndUpdateData("SiLookupTypeId").(int)
	result.Type = &metadataType

	// apis
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<RuntimeApiMetadataV15>").([]interface{}), &result.Apis)
	for index, api := range result.Apis {
		for methodIndex, method := range api.Methods {
			result.Apis[index].Methods[methodIndex].Outputs = metadataSiType[method.OutputsId]
			for inputIndex, arg := range method.Inputs {
				result.Apis[index].Methods[methodIndex].Inputs[inputIndex].Type = metadataSiType[arg.TypeId]
			}
		}
	}

	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("OuterEnumsMetadataV15"), &result.OuterEnums)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("CustomMetadataV15"), &result.Customer)
	m.Value = result
}

type MetadataV15Module struct {
	MetadataV14Module
	Docs []string `json:"docs"`
}

func (m *MetadataV15Module) Process() {
	m.MetadataV14Module.Process()
	result := MetadataV15Module{}
	_ = utiles.UnmarshalAny(m.Value, &result)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &result.Docs)
	m.Value = result
}

type RuntimeApiMetadataV15 struct {
	ScaleDecoder
}

// Process
//
//	"RuntimeApiMetadataV15": {
//	   "name": "Text",
//	   "methods": "Vec<RuntimeApiMethodMetadataV15>",
//	   "docs": "Vec<Text>"
//	 }
func (m *RuntimeApiMetadataV15) Process() {
	runtimeApiMetadata := RuntimeApiMetadata{}
	runtimeApiMetadata.Name = m.ProcessAndUpdateData("Text").(string)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<RuntimeApiMethodMetadataV15>").([]interface{}), &runtimeApiMetadata.Methods)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &runtimeApiMetadata.Docs)
	m.Value = runtimeApiMetadata
}

type RuntimeApiMethodParamMetadataV15 struct {
	ScaleDecoder
}

// Process
//
//	"RuntimeApiMethodParamMetadataV15": {
//	   "name": "Text",
//	   "type": "SiLookupTypeId"
//	 },
func (r *RuntimeApiMethodParamMetadataV15) Process() {
	ra := RuntimeApiMethodParamMetadata{}
	ra.Name = r.ProcessAndUpdateData("Text").(string)
	ra.TypeId = r.ProcessAndUpdateData("SiLookupTypeId").(int)
	r.Value = ra
}

type RuntimeApiMethodMetadataV15 struct {
	ScaleDecoder
}

func (r *RuntimeApiMethodMetadataV15) Process() {
	ra := RuntimeApiMethodMetadata{}
	ra.Name = r.ProcessAndUpdateData("Text").(string)

	_ = utiles.UnmarshalAny(r.ProcessAndUpdateData("Vec<RuntimeApiMethodParamMetadataV15>").([]interface{}), &ra.Inputs)
	ra.OutputsId = r.ProcessAndUpdateData("SiLookupTypeId").(int)
	_ = utiles.UnmarshalAny(r.ProcessAndUpdateData("Vec<Text>").([]interface{}), &ra.Docs)
	r.Value = ra
}

type OuterEnumsMetadataV15 struct{ ScaleDecoder }

func (o *OuterEnumsMetadataV15) Process() {
	ra := OuterEnumsMetadata{}
	ra.CallType = o.ProcessAndUpdateData("SiLookupTypeId").(int)
	ra.EventType = o.ProcessAndUpdateData("SiLookupTypeId").(int)
	ra.ErrorType = o.ProcessAndUpdateData("SiLookupTypeId").(int)
	// fmt.Println(ra.CallType, ra.ErrorType, ra.EventType)
	o.Value = ra
}

type CustomMetadataV15 struct{ ScaleDecoder }

func (c *CustomMetadataV15) Process() {
	ra := CustomMetadata{}
	mapping := c.ProcessAndUpdateData("BTreeMap<Text, CustomValueMetadata15>").([]interface{})
	_ = utiles.UnmarshalAny(mapping, &ra.Map)
	c.Value = ra
}

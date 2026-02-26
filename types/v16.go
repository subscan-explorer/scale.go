package types

import (
	"encoding/json"
	"fmt"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/utiles"
)

type MetadataV16Decoder struct {
	ScaleDecoder
}

// "MetadataV16": {
//    "lookup": "PortableRegistry",
//    "pallets": "Vec<PalletMetadataV16>",
//    "extrinsic": "ExtrinsicMetadataV16",
//    "apis": "Vec<RuntimeApiMetadataV16>"
//    "outerEnums": "OuterEnums15"
//    "custom": CustomMetadata15
//  }

func (m *MetadataV16Decoder) Process() {
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
	metadataV16ModuleCall := m.ProcessAndUpdateData("Vec<MetadataV16Module>").([]interface{})
	bm, _ := json.Marshal(metadataV16ModuleCall)

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
				if dep, ok := module.CallsDeprecationInfo[variant.Index]; ok {
					deprecation := dep
					call.DeprecationInfo = &deprecation
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
				if dep, ok := module.EventsDeprecationInfo[variant.Index]; ok {
					deprecation := dep
					event.DeprecationInfo = &deprecation
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
				if dep, ok := module.ErrorsDeprecationInfo[variant.Index]; ok {
					deprecation := dep
					moduleErr.DeprecationInfo = &deprecation
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
		for index, associated := range module.AssociatedTypes {
			modulesType[k].AssociatedTypes[index].Type = metadataSiType[associated.TypeId]
		}
		for index, view := range module.ViewFunctions {
			modulesType[k].ViewFunctions[index].Output = metadataSiType[view.OutputId]
			for inputIndex, input := range view.Inputs {
				modulesType[k].ViewFunctions[index].Inputs[inputIndex].Type = metadataSiType[input.TypeId]
			}
		}
	}
	result.Metadata.Modules = modulesType
	result.Extrinsic = decodeExtrinsicMetadataV16(&m.ScaleDecoder)

	for index, extension := range result.Extrinsic.SignedExtensions {
		result.Extrinsic.SignedExtensions[index].TypeString = metadataSiType[extension.Type]
		result.Extrinsic.SignedIdentifier = append(result.Extrinsic.SignedIdentifier, extension.Identifier)
	}
	for index, extension := range result.Extrinsic.TransactionExtensions {
		result.Extrinsic.TransactionExtensions[index].TypeString = metadataSiType[extension.Type]
		result.Extrinsic.TransactionExtensions[index].ImplicitString = metadataSiType[extension.Implicit]
	}

	registerOriginCaller(originCallers)

	// apis
	apiCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < apiCount; i++ {
		api := RuntimeApiMetadata{}
		api.Name = m.ProcessAndUpdateData("Text").(string)

		methodCount := m.ProcessAndUpdateData("Compact<u32>").(int)
		for j := 0; j < methodCount; j++ {
			method := RuntimeApiMethodMetadata{}
			method.Name = m.ProcessAndUpdateData("Text").(string)

			inputCount := m.ProcessAndUpdateData("Compact<u32>").(int)
			for k := 0; k < inputCount; k++ {
				input := RuntimeApiMethodParamMetadata{
					Name:   m.ProcessAndUpdateData("Text").(string),
					TypeId: m.ProcessAndUpdateData("SiLookupTypeId").(int),
				}
				input.Type = metadataSiType[input.TypeId]
				method.Inputs = append(method.Inputs, input)
			}

			method.OutputsId = m.ProcessAndUpdateData("SiLookupTypeId").(int)
			method.Outputs = metadataSiType[method.OutputsId]
			_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &method.Docs)
			method.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
			api.Methods = append(api.Methods, method)
		}

		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &api.Docs)
		api.Version = m.ProcessAndUpdateData("Compact<u32>").(int)
		api.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
		result.Apis = append(result.Apis, api)
	}

	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("OuterEnumsMetadataV15"), &result.OuterEnums)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("CustomMetadataV15"), &result.Customer)
	m.Value = result
}

type MetadataV16Module struct {
	ScaleDecoder
}

func (m *MetadataV16Module) Process() {
	result := MetadataModules{}
	result.Name = m.ProcessAndUpdateData("String").(string)

	// storage
	hasStorage := m.ProcessAndUpdateData("bool").(bool)
	if hasStorage {
		result.Prefix = m.ProcessAndUpdateData("String").(string)
		itemCount := m.ProcessAndUpdateData("Compact<u32>").(int)
		for i := 0; i < itemCount; i++ {
			storage := m.ProcessAndUpdateData("MetadataV14ModuleStorageEntry").(MetadataStorage)
			storage.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
			result.Storage = append(result.Storage, storage)
		}
	}

	// call
	hasCalls := m.ProcessAndUpdateData("bool").(bool)
	if hasCalls {
		var callsValue PalletLookUp
		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("PalletCallMetadataV14"), &callsValue)
		result.CallsValue = &callsValue
		result.CallsDeprecationInfo = decodeEnumDeprecationInfoV16(&m.ScaleDecoder)
	}

	// event
	hasEvents := m.ProcessAndUpdateData("bool").(bool)
	if hasEvents {
		var eventsValue PalletLookUp
		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("PalletEventMetadataV14"), &eventsValue)
		result.EventsValue = &eventsValue
		result.EventsDeprecationInfo = decodeEnumDeprecationInfoV16(&m.ScaleDecoder)
	}

	// constant
	constantCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < constantCount; i++ {
		constant := m.ProcessAndUpdateData("PalletConstantMetadataV14").(MetadataConstants)
		constant.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
		result.Constants = append(result.Constants, constant)
	}

	// error
	hasError := m.ProcessAndUpdateData("bool").(bool)
	if hasError {
		var errorsValue PalletLookUp
		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("PalletErrorMetadataV14"), &errorsValue)
		result.ErrorsValue = &errorsValue
		result.ErrorsDeprecationInfo = decodeEnumDeprecationInfoV16(&m.ScaleDecoder)
	}

	// associatedTypes
	associatedTypeCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < associatedTypeCount; i++ {
		associatedType := PalletAssociatedTypeMetadata{
			Name:   m.ProcessAndUpdateData("Text").(string),
			TypeId: m.ProcessAndUpdateData("SiLookupTypeId").(int),
		}
		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &associatedType.Docs)
		result.AssociatedTypes = append(result.AssociatedTypes, associatedType)
	}

	// viewFunctions
	viewFunctionCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < viewFunctionCount; i++ {
		viewFunction := PalletViewFunctionMetadata{
			ID:   m.ProcessAndUpdateData("[u8; 8]").(string),
			Name: m.ProcessAndUpdateData("Text").(string),
		}
		inputCount := m.ProcessAndUpdateData("Compact<u32>").(int)
		for j := 0; j < inputCount; j++ {
			viewFunction.Inputs = append(viewFunction.Inputs, RuntimeApiMethodParamMetadata{
				Name:   m.ProcessAndUpdateData("Text").(string),
				TypeId: m.ProcessAndUpdateData("SiLookupTypeId").(int),
			})
		}
		viewFunction.OutputId = m.ProcessAndUpdateData("SiLookupTypeId").(int)
		_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &viewFunction.Docs)
		viewFunction.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
		result.ViewFunctions = append(result.ViewFunctions, viewFunction)
	}

	result.Index = m.ProcessAndUpdateData("U8").(int)
	_ = utiles.UnmarshalAny(m.ProcessAndUpdateData("Vec<Text>").([]interface{}), &result.Docs)
	result.DeprecationInfo = decodeItemDeprecationInfoV16(&m.ScaleDecoder)
	m.Value = result
}

func decodeExtrinsicMetadataV16(m *ScaleDecoder) *ExtrinsicMetadata {
	result := &ExtrinsicMetadata{}
	result.Versions = m.ProcessAndUpdateData("Bytes").(string)
	result.AddressType = m.ProcessAndUpdateData("SiLookupTypeId").(int)
	result.CallType = m.ProcessAndUpdateData("SiLookupTypeId").(int)
	result.SignatureType = m.ProcessAndUpdateData("SiLookupTypeId").(int)

	byVersionCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < byVersionCount; i++ {
		txByVersion := TransactionExtensionsByVersion{Version: m.ProcessAndUpdateData("u8").(int)}
		versionExtCount := m.ProcessAndUpdateData("Compact<u32>").(int)
		for j := 0; j < versionExtCount; j++ {
			txByVersion.Extensions = append(txByVersion.Extensions, m.ProcessAndUpdateData("Compact<u32>").(int))
		}
		result.TransactionExtensionsByVersion = append(result.TransactionExtensionsByVersion, txByVersion)
	}

	txExtensionCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	for i := 0; i < txExtensionCount; i++ {
		txExtension := TransactionExtensionMetadata{
			Identifier: m.ProcessAndUpdateData("Text").(string),
			Type:       m.ProcessAndUpdateData("SiLookupTypeId").(int),
			Implicit:   m.ProcessAndUpdateData("SiLookupTypeId").(int),
		}
		result.TransactionExtensions = append(result.TransactionExtensions, txExtension)
		result.SignedExtensions = append(result.SignedExtensions, SignedExtensions{
			Identifier:       txExtension.Identifier,
			Type:             txExtension.Type,
			AdditionalSigned: txExtension.Implicit,
		})
	}
	return result
}

func decodeItemDeprecationInfoV16(m *ScaleDecoder) *DeprecationInfo {
	info := &DeprecationInfo{}
	switch m.ProcessAndUpdateData("u8").(int) {
	case 0:
		info.Type = "NotDeprecated"
	case 1:
		info.Type = "DeprecatedWithoutNote"
	case 2:
		info.Type = "Deprecated"
		info.Note = m.ProcessAndUpdateData("Text").(string)
		info.Since = decodeOptionalText(m)
	default:
		info.Type = "Unknown"
	}
	return info
}

func decodeEnumDeprecationInfoV16(m *ScaleDecoder) map[int]DeprecationInfo {
	itemCount := m.ProcessAndUpdateData("Compact<u32>").(int)
	if itemCount == 0 {
		return nil
	}
	result := make(map[int]DeprecationInfo, itemCount)
	for i := 0; i < itemCount; i++ {
		itemIndex := m.ProcessAndUpdateData("u8").(int)
		result[itemIndex] = decodeVariantDeprecationInfoV16(m)
	}
	return result
}

func decodeVariantDeprecationInfoV16(m *ScaleDecoder) DeprecationInfo {
	info := DeprecationInfo{}
	switch m.ProcessAndUpdateData("u8").(int) {
	case 0:
		info.Type = "DummyVariant"
	case 1:
		info.Type = "DeprecatedWithoutNote"
	case 2:
		info.Type = "Deprecated"
		info.Note = m.ProcessAndUpdateData("Text").(string)
		info.Since = decodeOptionalText(m)
	default:
		info.Type = "Unknown"
	}
	return info
}

func decodeOptionalText(m *ScaleDecoder) *string {
	value := m.ProcessAndUpdateData("Option<Text>")
	if value == nil {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		return nil
	}
	return &text
}

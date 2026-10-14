package scalecodec

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	scaleType "github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/blake2b"
)

type ExtrinsicParam struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	TypeName string      `json:"type_name"`
	Value    interface{} `json:"value"`
}

type ExtrinsicDecoder struct {
	scaleType.ScaleDecoder
	ExtrinsicLength     int                         `json:"extrinsic_length"`
	ExtrinsicHash       string                      `json:"extrinsic_hash"`
	VersionInfo         string                      `json:"version_info"`
	ContainsTransaction bool                        `json:"contains_transaction"`
	Address             interface{}                 `json:"address"`
	Signature           string                      `json:"signature"`
	Nonce               int                         `json:"nonce"`
	Era                 string                      `json:"era"`
	CallIndex           string                      `json:"call_index"`
	Params              []ExtrinsicParam            `json:"params"`
	ParamsRaw           string                      `json:"params_raw"`
	Metadata            *scaleType.MetadataStruct   `json:"-"`
	SignedExtensions    []scaleType.SignedExtension `json:"signed_extensions"`
	AdditionalCheck     []string
}

// https://github.com/polkadot-js/api/blob/master/packages/types/src/extrinsic/signedExtensions/index.ts#L24
var signedExts = map[string]bool{
	"CheckSpecVersion":         false,
	"CheckTxVersion":           false,
	"CheckGenesis":             false,
	"CheckMortality":           false,
	"CheckNonce":               true,
	"CheckWeight":              false,
	"ChargeTransactionPayment": true,
	"CheckBlockGasLimit":       false,
	"ChargeAssetTxPayment":     true,
	"CheckMetadataHash":        true,
}

func (e *ExtrinsicDecoder) Init(data scaleBytes.ScaleBytes, option *scaleType.ScaleDecoderOption) {
	if option == nil || option.Metadata == nil {
		panic("ExtrinsicDecoder option metadata required")
	}
	e.Params = []ExtrinsicParam{}
	e.Metadata = option.Metadata
	e.SignedExtensions = option.SignedExtensions
	e.AdditionalCheck = option.AdditionalCheck
	e.ScaleDecoder.Init(data, option)
}

func blake2_256(data []byte) string {
	checksum, _ := blake2b.New(32, []byte{})
	_, _ = checksum.Write(data)
	h := checksum.Sum(nil)
	return utiles.BytesToHex(h)
}

func (e *ExtrinsicDecoder) generateHash() string {
	var extrinsicData []byte
	// Data includes the preamble for every format, including general V5.
	if e.ExtrinsicLength > 0 {
		extrinsicData = e.Data.Data
	} else {
		extrinsicLengthType := scaleType.CompactU32{}
		extrinsicLengthType.Encode(len(e.Data.Data))
		extrinsicData = append(extrinsicLengthType.Data.Data[:], e.Data.Data[:]...)
	}
	return blake2_256(extrinsicData)
}

func signedExtensionsWithExtra(signedExtensions []scaleType.SignedExtensions) []scaleType.SignedExtensions {
	filtered := make([]scaleType.SignedExtensions, 0, len(signedExtensions))
	for _, ext := range signedExtensions {
		if ext.TypeString == "NULL" {
			continue
		}
		filtered = append(filtered, ext)
	}
	return filtered
}

func extractTip(value interface{}) decimal.Decimal {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, item := range typed {
			if strings.EqualFold(key, "tip") {
				return utiles.DecimalFromInterface(item)
			}
		}
	case nil:
		return decimal.Zero
	default:
		return utiles.DecimalFromInterface(typed)
	}
	return decimal.Zero
}

func extractNonce(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case uint8:
		return int(typed), true
	case uint16:
		return int(typed), true
	case uint32:
		return int(typed), true
	case uint64:
		return int(typed), true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case decimal.Decimal:
		return int(typed.IntPart()), true
	}
	return 0, false
}

func encodeEraValue(value interface{}, ext scaleType.SignedExtensions, option *scaleType.ScaleDecoderOption) string {
	if raw, ok := value.(string); ok {
		return raw
	}
	return utiles.TrimHex(scaleType.EncodeWithOpt(ext.TypeString, value, option))
}

func hasSignedExtensionValue(value interface{}) bool {
	if value == nil {
		return false
	}
	if typed, ok := value.(map[string]interface{}); ok && len(typed) == 0 {
		return false
	}
	return true
}

func shouldEncodeSignedExtension(identifier string, value interface{}) bool {
	switch identifier {
	case "CheckMortality", "CheckNonce", "ChargeTransactionPayment", "ChargeAssetTxPayment":
		return false
	}
	return hasSignedExtensionValue(value)
}

func shouldRecordSignedExtension(identifier string, value interface{}) bool {
	switch identifier {
	case "CheckMortality", "CheckNonce", "ChargeTransactionPayment":
		return false
	}
	return hasSignedExtensionValue(value)
}

// Metadata v14 describes each byte-bearing signed extension separately. Decode
// those types directly instead of relying on a process-global Extra tuple.
func (e *ExtrinsicDecoder) decodeExtrinsicExtra(result *GenericExtrinsic) bool {
	if e.Metadata.MetadataVersion < 14 || e.Metadata.Extrinsic == nil {
		return false
	}

	result.SignedExtensions = make(map[string]interface{})
	extensions := signedExtensionsWithExtra(e.Metadata.Extrinsic.SignedExtensions)
	for _, ext := range extensions {
		value := e.ProcessAndUpdateData(ext.TypeString)
		switch ext.Identifier {
		case "CheckMortality":
			e.Era = encodeEraValue(value, ext, &scaleType.ScaleDecoderOption{Metadata: e.Metadata, Spec: e.Spec})
		case "CheckNonce":
			if nonce, ok := extractNonce(value); ok {
				e.Nonce = nonce
			}
		case "ChargeTransactionPayment", "ChargeAssetTxPayment":
			result.Tip = extractTip(value)
		}
		if shouldRecordSignedExtension(ext.Identifier, value) {
			result.SignedExtensions[ext.Identifier] = value
		}
	}
	return true
}

// General extrinsics carry authorization inside the extension pipeline rather
// than a signed preamble. Preserve every extension for lossless re-encoding and
// expose the common fields to callers that index accounts and transaction fees.
func (e *ExtrinsicDecoder) decodeGeneralExtensions(result *GenericExtrinsic) {
	extensions, err := e.Metadata.Extrinsic.ExtensionsForVersion(result.TransactionExtensionVersion)
	if err != nil {
		panic(fmt.Errorf("decode general extrinsic: %w", err))
	}
	result.SignedExtensions = make(map[string]interface{}, len(extensions))
	for _, ext := range extensions {
		start := e.Data.Offset
		value := e.ProcessAndUpdateData(ext.TypeString)
		result.SignedExtensions[ext.Identifier] = value
		switch ext.Identifier {
		case "CheckMortality":
			result.Era = utiles.BytesToHex(e.Data.Data[start:e.Data.Offset])
			e.Era = result.Era
		case "CheckNonce":
			if nonce, ok := extractNonce(value); ok {
				result.Nonce = nonce
				e.Nonce = nonce
			}
		case "ChargeTransactionPayment", "ChargeAssetTxPayment":
			result.Tip = extractTip(value)
		case "VerifyMultiSignature":
			variants, ok := value.(map[string]interface{})
			if !ok {
				continue // Disabled authorization has no account or signature.
			}
			signed, ok := variants["Signed"].(map[string]interface{})
			if !ok {
				continue
			}
			result.AccountId = signed["account"]
			e.Address = result.AccountId
			result.Signer = signed["account"]
			result.AddressType = "AccountId"
			result.SignatureRaw = signed["signature"]
			if signature, ok := result.SignatureRaw.(map[string]interface{}); ok {
				for _, raw := range signature {
					if text, ok := raw.(string); ok {
						result.Signature = text
						e.Signature = text
					}
				}
			}
		}
	}
}

type GenericExtrinsic struct {
	VersionInfo                 string                 `json:"version_info"`
	ExtrinsicLength             int                    `json:"extrinsic_length"`
	AddressType                 string                 `json:"address_type"`
	Tip                         decimal.Decimal        `json:"tip"`
	SignedExtensions            map[string]interface{} `json:"signed_extensions"`
	AccountId                   interface{}            `json:"account_id"`
	Signer                      interface{}            `json:"signer"`        // map[string]interface or string
	Signature                   string                 `json:"signature"`     // Legacy string/enum payload, or full composite SCALE hex without 0x.
	SignatureRaw                interface{}            `json:"signature_raw"` // Original metadata-decoded value, retained for encoding.
	Nonce                       int                    `json:"nonce"`
	Era                         string                 `json:"era"`
	ExtrinsicHash               string                 `json:"extrinsic_hash"`
	CallModuleFunction          string                 `json:"call_module_function"`
	CallCode                    string                 `json:"call_code"`
	CallModule                  string                 `json:"call_module"`
	Params                      []ExtrinsicParam       `json:"params"`
	ParamsRaw                   string                 `json:"params_raw"`
	TransactionExtensionVersion int                    `json:"transaction_extension_version"`
}

// decodeSignature preserves the metadata-decoded value and records the exact
// composite span. Only actual enum string payloads use the legacy flat output;
// a struct with one string field must still retain its complete encoding.
func (e *ExtrinsicDecoder) decodeSignature() interface{} {
	start := e.Data.Offset
	raw := e.ProcessAndUpdateData("ExtrinsicSignature")
	span := e.Data.Data[start:e.Data.Offset]
	e.Signature = utiles.BytesToHex(span)
	if value, ok := raw.(string); ok {
		e.Signature = value
		return raw
	}
	runtimeType := scaleType.RuntimeType{Module: e.Module}
	codec, _, err := runtimeType.GetCodec("ExtrinsicSignature", e.Spec)
	if err != nil {
		panic(err)
	}
	if enum, ok := codec.(*scaleType.Enum); ok && enum.TypeMapping != nil && len(span) > 0 {
		index := int(span[0])
		if fields, ok := raw.(map[string]interface{}); ok && index < len(enum.TypeMapping.Names) {
			if value, ok := fields[enum.TypeMapping.Names[index]].(string); ok {
				e.Signature = value
			}
		}
	}
	return raw
}

func (e *ExtrinsicDecoder) Process() {
	e.ExtrinsicLength = e.ProcessAndUpdateData("Compact<u32>").(int)
	if e.ExtrinsicLength != e.Data.GetRemainingLength() {
		e.ExtrinsicLength = 0
		e.Data.Reset()
	}

	e.VersionInfo = utiles.BytesToHex(e.NextBytes(1))

	e.ContainsTransaction = utiles.U256(e.VersionInfo).Int64() >= 80

	result := GenericExtrinsic{
		ExtrinsicLength: e.ExtrinsicLength,
		VersionInfo:     e.VersionInfo,
	}
	if e.VersionInfo == "04" || e.VersionInfo == "84" {
		decodedExtrinsicExtra := false
		if e.ContainsTransaction {
			// Address
			result.Signer = e.ProcessAndUpdateData(utiles.TrueOrElse(e.Metadata.MetadataVersion >= 14 && scaleType.HasReg("ExtrinsicSigner"), "ExtrinsicSigner", "Address"))
			switch v := result.Signer.(type) {
			case string:
				e.Address = v
				result.AddressType = "AccountId"
			case map[string]interface{}:
				for name, value := range v {
					result.AddressType = name
					e.Address = value
				}
			}
			// ExtrinsicSignature
			result.SignatureRaw = e.decodeSignature()
			decodedExtrinsicExtra = e.decodeExtrinsicExtra(&result)
			if !decodedExtrinsicExtra {
				e.Era = e.ProcessAndUpdateData("EraExtrinsic").(string)
				if e.Metadata.MetadataVersion < 14 || utiles.SliceIndex("CheckNonce", e.Metadata.Extrinsic.SignedIdentifier) < 0 {
					e.Nonce = int(e.ProcessAndUpdateData("Compact<U64>").(uint64))
				}
			}
			if e.Metadata.MetadataVersion < 14 {
				// confirm metadata extrinsic has ChargeTransactionPayment
				if e.Metadata.Extrinsic != nil {
					if utiles.SliceIndex("ChargeTransactionPayment", e.Metadata.Extrinsic.SignedIdentifier) != -1 {
						result.Tip = utiles.DecimalFromInterface(e.ProcessAndUpdateData("Compact<Balance>"))
					}
				} else {
					result.Tip = utiles.DecimalFromInterface(e.ProcessAndUpdateData("Compact<Balance>"))
				}
			}
			// spec SignedExtensions
			if len(result.SignedExtensions) == 0 {
				result.SignedExtensions = make(map[string]interface{})
			}
			if len(e.SignedExtensions) > 0 {
				for _, extension := range e.SignedExtensions {
					if utiles.SliceIndex(extension.Name, e.Metadata.Extrinsic.SignedIdentifier) != -1 {
						for _, v := range extension.AdditionalSigned {
							result.SignedExtensions[v.Name] = e.ProcessAndUpdateData(v.Type)
						}
					}
				}
			} else if e.Metadata.MetadataVersion >= 14 && !decodedExtrinsicExtra && !scaleType.HasReg("ExtrinsicExtra") {
				for _, ext := range e.Metadata.Extrinsic.SignedExtensions {
					if enable := signedExts[ext.Identifier]; enable || utiles.SliceIndex(ext.Identifier, e.AdditionalCheck) >= 0 {
						if ext.Identifier == "ChargeTransactionPayment" {
							result.Tip = utiles.DecimalFromInterface(e.ProcessAndUpdateData("Compact<Balance>"))
						} else if ext.Identifier == "CheckNonce" {
							e.Nonce = int(e.ProcessAndUpdateData("Compact<U64>").(uint64))
						} else {
							result.SignedExtensions[ext.Identifier] = e.ProcessAndUpdateData(ext.TypeString)
						}
					}
				}
			}
			e.ExtrinsicHash = e.generateHash()
		}
		e.CallIndex = utiles.BytesToHex(e.NextBytes(2))
	} else if e.VersionInfo == "05" || e.VersionInfo == "85" {
		if e.ContainsTransaction {
			// Address
			result.Signer = e.ProcessAndUpdateData(utiles.TrueOrElse(e.Metadata.MetadataVersion >= 14 && scaleType.HasReg("ExtrinsicSigner"), "ExtrinsicSigner", "Address"))
			switch v := result.Signer.(type) {
			case string:
				e.Address = v
				result.AddressType = "AccountId"
			case map[string]interface{}:
				for name, value := range v {
					result.AddressType = name
					e.Address = value
				}
			}
			// ExtrinsicSignature
			result.SignatureRaw = e.decodeSignature()
			result.TransactionExtensionVersion = e.ProcessAndUpdateData("U8").(int)
			e.Era = e.ProcessAndUpdateData("EraExtrinsic").(string)
			e.Nonce = int(e.ProcessAndUpdateData("Compact<U64>").(uint64))
			// spec SignedExtensions
			result.SignedExtensions = make(map[string]interface{})
			for _, ext := range e.Metadata.Extrinsic.SignedExtensions {
				if enable := signedExts[ext.Identifier]; enable || utiles.SliceIndex(ext.Identifier, e.AdditionalCheck) >= 0 {
					if ext.Identifier == "ChargeTransactionPayment" {
						result.Tip = utiles.DecimalFromInterface(e.ProcessAndUpdateData("Compact<Balance>"))
					} else if ext.Identifier == "CheckNonce" {
						e.Nonce = int(e.ProcessAndUpdateData("Compact<U64>").(uint64))
					} else {
						result.SignedExtensions[ext.Identifier] = e.ProcessAndUpdateData(ext.TypeString)
					}
				}
			}
			e.ExtrinsicHash = e.generateHash()
		}
		e.CallIndex = utiles.BytesToHex(e.NextBytes(2))
	} else if e.VersionInfo == "45" {
		e.ExtrinsicHash = e.generateHash()
		result.TransactionExtensionVersion = e.ProcessAndUpdateData("U8").(int)
		e.decodeGeneralExtensions(&result)
		e.CallIndex = utiles.BytesToHex(e.NextBytes(2))
	} else {
		panic(fmt.Sprintf("Extrinsics version %s is not support", e.VersionInfo))
	}
	if e.CallIndex == "" {
		panic("Not find Extrinsic Lookup, please check type registry")
	}

	call, ok := e.Metadata.CallIndex[e.CallIndex]
	if !ok {
		panic(fmt.Sprintf("Not find Extrinsic Lookup %s, please check metadata info", e.CallIndex))
	}
	e.Module = call.Module.Name
	offset := e.Data.Offset
	for _, arg := range call.Call.Args {
		e.Params = append(e.Params, ExtrinsicParam{Name: arg.Name, Type: arg.Type, Value: e.ProcessAndUpdateData(arg.Type), TypeName: arg.TypeName})
	}
	e.ParamsRaw = utiles.BytesToHex(e.Data.Data[offset:e.Data.Offset])

	if e.ContainsTransaction {
		result.AccountId = e.Address
		result.Signature = e.Signature
		result.Nonce = e.Nonce
		result.Era = e.Era
	}
	result.ExtrinsicHash = utiles.AddHex(e.ExtrinsicHash)
	result.CallCode = e.CallIndex
	result.CallModuleFunction = call.Call.Name
	result.CallModule = call.Module.Name
	result.Params = e.Params
	result.ParamsRaw = e.ParamsRaw
	e.Value = &result
}

/*
Encode extrinsic with option
opt.Metadata is required
return hex string, if error, return empty string

Example:
m := scalecodec.MetadataDecoder{}
m.Init(utiles.HexToBytes(Kusama9370))
_ = m.Process()
option := types.ScaleDecoderOption{Metadata: &m.Metadata}
genericExtrinsic := scalecodec.GenericExtrinsic{
	VersionInfo:  "84",
	CallCode:     "0400",
	Nonce:        0,
	Era:          "00",
	Signer:       map[string]interface{}{"Id": "0xe673cb35ffaaf7ab98c4e9268bfa9b4a74e49d41c8225121c346db7a7dd06d88"},
	SignatureRaw: map[string]interface{}{"Ed25519": "0xfce9453b1442bba86c2781e755a29c8a215ccf4b65ce81eeaa5b5a04dcdb79a54525cc86969f910c71c05f84aeab9c205022ecd4aa2abb4a3c3667f09dd16e0b"},
	Params: []scalecodec.ExtrinsicParam{
		{Value: map[string]interface{}{"Id": "0x0770e0831a275b534f7507c8ebd9f5f982a55053c9dc672da886ef41a6b5c628"}}, {Value: "1094000000000"},
	},
}
fmt.Println(genericExtrinsic.Encode(&option))
*/

func (g *GenericExtrinsic) Encode(opt *scaleType.ScaleDecoderOption) (string, error) {
	if opt.Metadata == nil {
		return "", errors.New("invalid metadata")
	}
	data := g.VersionInfo
	if g.VersionInfo == "84" {
		data = data + scaleType.Encode(utiles.TrueOrElse(opt.Metadata.MetadataVersion >= 14 && scaleType.HasReg("ExtrinsicSigner"), "ExtrinsicSigner", "AccountId"), g.Signer) // accountId
		data = data + scaleType.Encode("ExtrinsicSignature", g.SignatureRaw)                                                                                                   // signature
		data = data + scaleType.Encode("EraExtrinsic", g.Era)                                                                                                                  // era
		data = data + scaleType.Encode("Compact<U64>", g.Nonce)                                                                                                                // nonce
		if len(opt.Metadata.Extrinsic.SignedIdentifier) > 0 && utiles.SliceIndex("ChargeTransactionPayment", opt.Metadata.Extrinsic.SignedIdentifier) > -1 {
			data = data + scaleType.Encode("Compact<Balance>", g.Tip) // tip
		}
		for _, ext := range opt.Metadata.Extrinsic.SignedExtensions {
			if extension, ok := g.SignedExtensions[ext.Identifier]; ok && shouldEncodeSignedExtension(ext.Identifier, extension) {
				data = data + scaleType.Encode(ext.TypeString, extension)
			}
		}
	} else if g.VersionInfo == "85" {
		data = data + scaleType.Encode(utiles.TrueOrElse(opt.Metadata.MetadataVersion >= 14 && scaleType.HasReg("ExtrinsicSigner"), "ExtrinsicSigner", "AccountId"), g.Signer)
		data = data + scaleType.Encode("ExtrinsicSignature", g.SignatureRaw)
		data = data + scaleType.Encode("U8", g.TransactionExtensionVersion)
		data = data + scaleType.Encode("EraExtrinsic", g.Era)
		data = data + scaleType.Encode("Compact<U64>", g.Nonce)
		for _, ext := range opt.Metadata.Extrinsic.SignedExtensions {
			if enable := signedExts[ext.Identifier]; enable || utiles.SliceIndex(ext.Identifier, opt.AdditionalCheck) >= 0 {
				if ext.Identifier == "ChargeTransactionPayment" {
					data = data + scaleType.Encode("Compact<Balance>", g.Tip)
				} else if ext.Identifier == "CheckNonce" {
					data = data + scaleType.Encode("Compact<U64>", g.Nonce)
				} else if extension, ok := g.SignedExtensions[ext.Identifier]; ok {
					data = data + scaleType.Encode(ext.TypeString, extension)
				}
			}
		}
	} else if g.VersionInfo == "45" {
		extensions, err := opt.Metadata.Extrinsic.ExtensionsForVersion(g.TransactionExtensionVersion)
		if err != nil {
			return "", fmt.Errorf("encode general extrinsic: %w", err)
		}
		data = data + scaleType.Encode("U8", g.TransactionExtensionVersion)
		for _, ext := range extensions {
			if extension, ok := g.SignedExtensions[ext.Identifier]; ok {
				data = data + scaleType.EncodeWithOpt(ext.TypeString, extension, opt)
			} else if ext.TypeString != "NULL" {
				return "", fmt.Errorf("encode general extrinsic: missing extension %s", ext.Identifier)
			}
		}
	}

	data = data + g.CallCode
	call, ok := opt.Metadata.CallIndex[g.CallCode]

	if !ok {
		return "", fmt.Errorf("not find Extrinsic Lookup %s, please check metadata info", g.CallCode)
	}

	if len(g.Params) != len(call.Call.Args) {
		return "", fmt.Errorf("extrinsic params length not match, expect %d, got %d", len(call.Call.Args), len(g.Params))
	}

	for index, arg := range call.Call.Args {
		data = data + utiles.TrimHex(scaleType.EncodeWithOpt(arg.Type, g.Params[index].Value, opt))
	}

	return scaleType.Encode("Compact<u32>", len(utiles.HexToBytes(data))) + data, nil
}

// ToMap GenericExtrinsic convert to map[string]interface
func (g *GenericExtrinsic) ToMap() map[string]interface{} {
	var r map[string]interface{}
	b, err := json.Marshal(g)
	if err != nil {
		return nil
	}
	_ = json.Unmarshal(b, &r)
	return r
}

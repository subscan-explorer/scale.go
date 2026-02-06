//go:generate go run ../tools/gen_registry

package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/types/override"
	"github.com/itering/scale.go/utiles"
)

type RuntimeType struct {
	Module string
}

type CodecFactory func() Decoder

type codecCacheEntry struct {
	factory CodecFactory
	subType string
}

type Special struct {
	Version  []int
	Registry CodecFactory
}

var (
	TypeRegistry        map[string]CodecFactory
	TypeRegistryLock    = &sync.RWMutex{}
	codecCache          = make(map[string]codecCacheEntry)
	codecCacheLock      = &sync.RWMutex{}
	specialRegistry     = make(map[string][]Special)
	specialRegistryLock = &sync.RWMutex{}
	V14Types            = make(map[string]source.TypeStruct)
	V14TypesLock        = &sync.RWMutex{}
)

func HasReg(typeName string) bool {
	TypeRegistryLock.RLock()
	_, ok := TypeRegistry[strings.ToLower(typeName)]
	TypeRegistryLock.RUnlock()
	return ok
}

// Clean all type registry
func Clean() {
	TypeRegistry = nil
	specialRegistry = make(map[string][]Special)
	V14Types = make(map[string]source.TypeStruct)
	regBaseType()
}

func init() {
	regBaseType()
}

func regBaseType() {
	registry := make(map[string]CodecFactory, len(baseCodecFactories)+24)
	for key, factory := range baseCodecFactories {
		registry[key] = factory
	}
	registry["compact<u32>"] = registry["compactu32"]
	registry["compact<moment>"] = func() Decoder { return &CompactMoment{} }
	registry["str"] = registry["string"]
	registry["hash"] = registry["h256"]
	registry["blockhash"] = registry["h256"]
	registry["i8"] = func() Decoder { return &IntFixed{FixedLength: 1} }
	registry["i16"] = func() Decoder { return &IntFixed{FixedLength: 2} }
	registry["i32"] = func() Decoder { return &IntFixed{FixedLength: 4} }
	registry["i64"] = func() Decoder { return &IntFixed{FixedLength: 8} }
	registry["i128"] = func() Decoder { return &IntFixed{FixedLength: 16} }
	registry["i256"] = func() Decoder { return &IntFixed{FixedLength: 32} }
	registry["h128"] = func() Decoder { return &FixedU8{FixedLength: 16} }
	registry["[u8; 32]"] = func() Decoder { return &FixedU8{FixedLength: 32} }
	registry["[u8; 64]"] = func() Decoder { return &FixedU8{FixedLength: 64} }
	registry["[u8; 65]"] = func() Decoder { return &FixedU8{FixedLength: 65} }
	registry["[u8; 16]"] = func() Decoder { return &FixedU8{FixedLength: 16} }
	registry["[u8; 20]"] = func() Decoder { return &FixedU8{FixedLength: 20} }
	registry["[u8; 8]"] = func() Decoder { return &FixedU8{FixedLength: 8} }
	registry["[u8; 4]"] = func() Decoder { return &FixedU8{FixedLength: 4} }
	registry["[u8; 2]"] = func() Decoder { return &FixedU8{FixedLength: 2} }
	registry["[u8; 256]"] = func() Decoder { return &FixedU8{FixedLength: 256} }
	registry["[u128; 3]"] = func() Decoder { return &FixedArray{FixedLength: 3, SubType: "u128"} }
	TypeRegistryLock.Lock()
	TypeRegistry = registry
	TypeRegistryLock.Unlock()
	resetCodecCache()
	// todo change load source pallet type to lazy load
	RegCustomTypes(source.LoadTypeRegistry([]byte(source.BaseType)))
}

func resetCodecCache() {
	codecCacheLock.Lock()
	codecCache = make(map[string]codecCacheEntry)
	codecCacheLock.Unlock()
}

func (r *RuntimeType) getCodecInstant(t string, spec int) (Decoder, CodecFactory, error) {
	t = override.ModuleType(strings.ToLower(t), r.Module)
	factory, err := r.specialVersionCodec(t, spec)

	if err != nil {
		TypeRegistryLock.RLock()
		factory = TypeRegistry[strings.ToLower(t)]
		TypeRegistryLock.RUnlock()
		// fixed array
		if factory == nil && t != "[]" && string(t[0]) == "[" && t[len(t)-1:] == "]" {
			if typePart := strings.Split(t[1:len(t)-1], ";"); len(typePart) >= 2 {
				remainPart := typePart[0 : len(typePart)-1]
				fixed := FixedArray{
					FixedLength: utiles.StringToInt(strings.TrimSpace(typePart[len(typePart)-1])),
					SubType:     strings.TrimSpace(strings.Join(remainPart, ";")),
				}
				factory = func() Decoder { return &fixed }
			}
		}
		if factory == nil {
			return nil, nil, errors.New("Scale codec type nil" + t)
		}
	}

	return factory(), factory, nil
}

func (r *RuntimeType) GetCodec(typeString string, spec int) (Decoder, string, error) {
	var typeParts []string
	typeString = convert.ConvertType(typeString)
	cacheKey := fmt.Sprintf("%s|%d|%s", r.Module, spec, typeString)
	codecCacheLock.RLock()
	entry, ok := codecCache[cacheKey]
	codecCacheLock.RUnlock()
	if ok {
		return entry.factory(), entry.subType, nil
	}

	// complex
	if typeString[len(typeString)-1:] == ">" {
		decoder, factory, err := r.getCodecInstant(typeString, spec)
		if err == nil {
			codecCacheLock.Lock()
			codecCache[cacheKey] = codecCacheEntry{factory: factory, subType: ""}
			codecCacheLock.Unlock()
			return decoder, "", nil
		}
		reg := regexp.MustCompile("^([^<]*)<(.+)>$")
		typeParts = reg.FindStringSubmatch(typeString)
	}

	if len(typeParts) > 0 {
		decoder, factory, err := r.getCodecInstant(typeParts[1], spec)
		if err == nil {
			codecCacheLock.Lock()
			codecCache[cacheKey] = codecCacheEntry{factory: factory, subType: typeParts[2]}
			codecCacheLock.Unlock()
			return decoder, typeParts[2], nil
		}
	} else {
		decoder, factory, err := r.getCodecInstant(typeString, spec)
		if err == nil {
			codecCacheLock.Lock()
			codecCache[cacheKey] = codecCacheEntry{factory: factory, subType: ""}
			codecCacheLock.Unlock()
			return decoder, "", nil
		}
	}

	// Tuple
	if typeString != "()" && string(typeString[0]) == "(" && typeString[len(typeString)-1:] == ")" {
		decoder, _, err := r.getCodecInstant("Struct", spec)
		if err != nil {
			return nil, "", err
		}
		s, ok := decoder.(*Struct)
		if !ok {
			return nil, "", fmt.Errorf("invalid struct decoder for %s", typeString)
		}
		s.TypeString = typeString
		s.buildStruct()
		codecCacheLock.Lock()
		codecCache[cacheKey] = codecCacheEntry{factory: func() Decoder {
			clone := Struct{}
			clone.TypeString = typeString
			clone.buildStruct()
			return &clone
		}, subType: ""}
		codecCacheLock.Unlock()
		return s, "", nil
	}

	// namespace
	if strings.Contains(typeString, "::") && typeString != "::" {
		namespaceSlice := strings.Split(typeString, "::")
		return r.GetCodec(namespaceSlice[len(namespaceSlice)-1], spec)
	}

	return nil, "", fmt.Errorf("scale codec type nil %s", typeString)
}

func (r *RuntimeType) specialVersionCodec(t string, spec int) (CodecFactory, error) {
	var factory CodecFactory
	specialRegistryLock.RLock()
	specials, ok := specialRegistry[t]
	specialRegistryLock.RUnlock()
	if ok {
		for _, special := range specials {
			if spec >= special.Version[0] && spec <= special.Version[1] {
				factory = special.Registry
				return factory, nil
			}
		}
	}
	return factory, fmt.Errorf("not found")
}

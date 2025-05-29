package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types/convert"
	"github.com/itering/scale.go/utiles"
)

func newStruct(names, typeString []string) *TypeMapping {
	if len(names) != len(typeString) {
		panic("init newStruct names and typeString length not equal")
	}
	if len(names) == 0 {
		return nil
	}
	return &TypeMapping{Names: names, Types: typeString}
}

func RegCustomTypes(registry map[string]source.TypeStruct) {
	for key := range registry {
		typeStruct := registry[key]
		if typeStruct.V14 {
			V14TypesLock.RLock()
			_, ok := V14Types[key]
			V14TypesLock.RUnlock()
			if !ok {
				V14TypesLock.Lock()
				V14Types[key] = typeStruct
				V14TypesLock.Unlock()
			} else {
				if typeStruct.SpecVec == 0 || EqTypeStringWithTypeStruct(key, &typeStruct) {
					return
				}
				// history spec type
				key = fmt.Sprintf("%s#%d-%d", key, typeStruct.SpecVec, typeStruct.SpecVec)
			}
		}
		key = strings.ToLower(key)
		switch typeStruct.Type {
		case "string":
			typeString := convert.ConvertType(typeStruct.TypeString)
			TypeRegistryLock.RLock()
			instant := TypeRegistry[strings.ToLower(typeString)]
			TypeRegistryLock.RUnlock()
			if instant != nil {
				regCustomKey(key, instant)
				continue
			}

			// Explained
			if explainedType, ok := registry[typeString]; ok {
				if explainedType.Type == "string" {
					TypeRegistryLock.RLock()
					instant = TypeRegistry[strings.ToLower(explainedType.TypeString)]
					TypeRegistryLock.RUnlock()
					if instant != nil {
						regCustomKey(key, instant)
						continue
					}
				} else {
					RegCustomTypes(map[string]source.TypeStruct{key: registry[typeString]})
				}
			}

			// subtype vec|option
			if typeString[len(typeString)-1:] == ">" {
				reg := regexp.MustCompile("^([^<]*)<(.+)>$")
				typeParts := reg.FindStringSubmatch(typeString)
				if len(typeParts) > 2 {
					if strings.EqualFold(typeParts[1], "vec") {
						v := Vec{}
						v.SubType = typeParts[2]
						regCustomKey(key, &v)
						continue
					} else if strings.EqualFold(typeParts[1], "option") {
						v := Option{}
						v.SubType = typeParts[2]
						regCustomKey(key, &v)
						continue
					} else if strings.EqualFold(typeParts[1], "compact") {
						v := Compact{}
						v.SubType = typeParts[2]
						regCustomKey(key, &v)
						continue
					} else if strings.EqualFold(typeParts[1], "BTreeMap") {
						v := BTreeMap{}
						v.SubType = typeParts[2]
						regCustomKey(key, &v)
						continue
					} else if strings.EqualFold(typeParts[1], "BTreeSet") {
						v := BTreeSet{}
						v.SubType = typeParts[2]
						regCustomKey(key, &v)
						continue
					}
				}

			}

			// Tuple
			if typeString != "()" && string(typeString[0]) == "(" && typeString[len(typeString)-1:] == ")" {
				s := Struct{}
				s.TypeString = typeString
				s.buildStruct()
				regCustomKey(key, &s)
				continue
			}

			// Array
			if typeString != "[]" && string(typeString[0]) == "[" && typeString[len(typeString)-1:] == "]" {
				if typePart := strings.Split(typeString[1:len(typeString)-1], ";"); len(typePart) == 2 {
					fixed := FixedArray{
						FixedLength: utiles.StringToInt(strings.TrimSpace(typePart[1])),
						SubType:     strings.TrimSpace(typePart[0]),
					}
					regCustomKey(key, &fixed)
					continue
				}
			}
			// result
			if strings.HasPrefix(typeString, "Result<") && strings.HasSuffix(typeString, ">") {
				reg := regexp.MustCompile("^Result<(.+)>$")
				typeParts := reg.FindStringSubmatch(typeString)
				if len(typeParts) > 1 {
					r := Result{}
					r.SubType = typeParts[1]
					regCustomKey(key, &r)
					continue
				}
			}
		case "struct":
			var names, typeStrings []string
			for _, v := range typeStruct.TypeMapping {
				names = append(names, v[0])
				typeStrings = append(typeStrings, v[1])
			}
			s := Struct{}
			s.TypeMapping = newStruct(names, typeStrings)

			regCustomKey(key, &s)
			continue
		case "enum":
			var names, typeStrings []string
			for _, v := range typeStruct.TypeMapping {
				names = append(names, v[0])
				typeStrings = append(typeStrings, v[1])
			}
			e := Enum{ValueList: typeStruct.ValueList}
			e.TypeMapping = newStruct(names, typeStrings)
			regCustomKey(key, &e)
			continue
		case "set":
			regCustomKey(key, &Set{ValueList: typeStruct.ValueList, BitLength: typeStruct.BitLength})
			continue
		}
	}
}

func regCustomKey(key string, rt interface{}) {
	slice := strings.Split(key, "#")
	if len(slice) == 2 { // for Special
		special := Special{Registry: rt, Version: []int{0, 99999999}}
		if version := strings.Split(slice[1], "-"); len(version) == 2 {
			special.Version[0] = utiles.StringToInt(version[0])
			if version[1] != "?" {
				special.Version[1] = utiles.StringToInt(version[1])
			}
		}
		specialRegistryLock.RLock()
		instant, ok := specialRegistry[slice[0]]
		specialRegistryLock.RUnlock()
		if ok {
			instant = append(instant, special)
		} else {
			instant = []Special{special}
		}
		specialRegistryLock.Lock()
		specialRegistry[slice[0]] = instant
		specialRegistryLock.Unlock()
	} else {
		TypeRegistryLock.Lock()
		TypeRegistry[key] = rt
		TypeRegistryLock.Unlock()
	}

}

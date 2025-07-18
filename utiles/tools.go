package utiles

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

func StringToInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0

}

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func AddHex(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return "0x" + s
}

func U256(v string) *big.Int {
	v = strings.TrimPrefix(v, "0x")
	bn := new(big.Int)
	n, _ := bn.SetString(v, 16)
	return n
}

func HexToBytes(s string) []byte {
	s = strings.TrimPrefix(s, "0x")
	c := make([]byte, hex.DecodedLen(len(s)))
	_, _ = hex.Decode(c, []byte(s))
	return c
}

func BytesToHex(b []byte) string {
	c := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(c, b)
	return string(c)
}

func IntToHex(i interface{}) string {
	return fmt.Sprintf("%x", i)
}

func UniqueSlice(s []string) (list []string) {
	keys := make(map[string]bool)
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return
}

func ReverseBytes(b []byte) []byte {
	a := make([]byte, len(b))
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = b[j], b[i]
	}
	return a
}

func ToString(i interface{}) string {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	default:
		b, _ := json.Marshal(i)
		val = string(b)
	}
	return val
}

func TrimHex(s string) string {
	return strings.TrimPrefix(s, "0x")
}

func BytesToBnHex(value []byte) string {
	var h string
	for _, b := range value {
		h += fmt.Sprintf("%d", int(b))
	}
	return h
}

func Debug(i interface{}) {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	case error:
		val = i.Error()
	default:
		b, _ := json.MarshalIndent(i, "", "  ")
		val = string(b)
	}
	fmt.Println(val)
}

func IsASCII(b []byte) bool {
	for _, c := range b {
		if c > 127 || (c < 32 && !IntInSlice(int(c), []int{9, 10, 13})) {
			return false
		}
	}
	return true
}

func SliceIndex(a string, list []string) int {
	for index, b := range list {
		if b == a {
			return index
		}
	}
	return -1
}

func TrueOrElse(expect bool, a, b string) string {
	if expect {
		return a
	}
	return b
}

func U8Encode(i int) string {
	bs := make([]byte, 1)
	bs[0] = byte(i)
	return BytesToHex(bs)
}

func IsNil(a interface{}) bool {
	if a == nil {
		return true
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Interface, reflect.Func:
		return reflect.ValueOf(a).IsNil()
	}
	return false
}

// GetEnumValue  get enum single key && value
func GetEnumValue(e map[string]interface{}) (string, interface{}, error) {
	for key, v := range e {
		return key, v, nil
	}
	return "", "", errors.New("empty enum")
}

// UnmarshalAny unmarshal any type
func UnmarshalAny(raw interface{}, r interface{}) error {
	switch v := raw.(type) {
	case string:
		if v == "" {
			return nil
		}

		return json.Unmarshal([]byte(v), &r)
	case []uint8:
		if len(v) == 0 {
			return nil
		}

		return json.Unmarshal(v, &r)
	default:
		b, _ := json.Marshal(v)
		return json.Unmarshal(b, r)
	}
}

func DecimalFromInterface(i interface{}) decimal.Decimal {
	switch i := i.(type) {
	case int:
		return decimal.New(int64(i), 0)
	case int64:
		return decimal.New(i, 0)
	case uint64:
		return decimal.New(int64(i), 0)
	case float64:
		return decimal.NewFromFloat(i)
	case string:
		r, _ := decimal.NewFromString(i)
		return r
	case decimal.Decimal:
		return i
	case *big.Int:
		return decimal.NewFromBigInt(i, 0)
	}
	return decimal.Zero
}

func U256DecoderToBigInt(u256 string) *big.Int {
	reverseData := ReverseBytes(HexToBytes(u256))
	return U256(BytesToHex(reverseData))
}

package types

import (
	"strings"

	"github.com/itering/scale.go/utiles"
)

type Bytes struct{ ScaleDecoder }

func (b *Bytes) Process() {
	length := b.ProcessAndUpdateData("Compact<u32>").(int)
	if value := b.NextBytes(length); utiles.IsASCII(value) {
		b.Value = string(value)
	} else {
		b.Value = utiles.AddHex(utiles.BytesToHex(value))
	}
}

func (b *Bytes) Encode(value interface{}) string {
	valueStr, ok := value.(string)
	if !ok {
		panic("invalid bytes input")
	}
	var bytes []byte
	if strings.HasPrefix(valueStr, "0x") {
		valueStr = utiles.TrimHex(valueStr)
		if len(valueStr)%2 == 1 {
			valueStr += "0"
		}
	} else {
		valueStr = utiles.BytesToHex([]byte(valueStr))
	}
	bytes = utiles.HexToBytes(valueStr)
	return Encode("Compact<u32>", len(bytes)) + valueStr
}

func (b *Bytes) TypeStructString() string {
	return "Bytes"
}

type HexBytes struct{ ScaleDecoder }

func (h *HexBytes) Process() {
	h.Value = utiles.AddHex(utiles.BytesToHex(h.NextBytes(h.ProcessAndUpdateData("Compact<u32>").(int))))
}

func (h *HexBytes) Encode(value interface{}) string {
	valueStr, ok := value.(string)
	if !ok {
		panic("invalid hexbytes input")
	}
	valueStr = utiles.TrimHex(valueStr)
	if len(valueStr)%2 == 1 {
		valueStr += "0"
	}
	bytes := utiles.HexToBytes(valueStr)
	return Encode("Compact<u32>", len(bytes)) + valueStr
}

func (h *HexBytes) TypeStructString() string {
	return "Bytes"
}

type String struct{ Bytes }

func (s *String) Encode(value interface{}) string {
	valueStr, ok := value.(string)
	if !ok {
		panic("invalid string input")
	}
	bytes := []byte(valueStr)
	return Encode("Compact<u32>", len(bytes)) + utiles.BytesToHex(bytes)
}

func (s *String) TypeStructString() string {
	return "String"
}

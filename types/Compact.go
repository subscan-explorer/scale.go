package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/shopspring/decimal"
)

type Compact struct {
	ScaleDecoder
	CompactLength int    `json:"compact_length"`
	CompactBytes  []byte `json:"compact_bytes"`
}

func (c *Compact) ProcessCompactBytes() {
	compactByte := c.NextBytes(1)
	var byteMod = 0
	if len(compactByte) != 0 {
		byteMod = int(compactByte[0]) % 4
	}
	switch byteMod {
	case 0:
		c.CompactLength = 1
		c.CompactBytes = compactByte
	case 1:
		c.CompactLength = 2
		c.CompactBytes = append(compactByte[:], c.NextBytes(c.CompactLength - 1)[:]...)
	case 2:
		c.CompactLength = 4
		c.CompactBytes = append(compactByte[:], c.NextBytes(c.CompactLength - 1)[:]...)
	default:
		c.CompactLength = 5 + (int(compactByte[0])-3)/4
		c.CompactBytes = c.NextBytes(c.CompactLength - 1)
	}
}

func (c *Compact) Process() {
	c.ProcessCompactBytes()
	if c.SubType == "" {
		c.Value = c.CompactBytes
		return
	}
	s := ScaleDecoder{TypeString: c.SubType, Data: scaleBytes.ScaleBytes{Data: c.CompactBytes}}
	byteData := s.ProcessAndUpdateData(c.SubType)
	var divValue uint64 = 1
	if c.CompactLength <= 4 {
		divValue = 4
	}
	switch v := byteData.(type) {
	case uint64:
		c.Value = v / divValue
	case uint16:
		c.Value = uint64(v) / divValue
	case uint32:
		c.Value = uint64(v) / divValue
	case int:
		c.Value = uint64(v) / divValue
	case decimal.Decimal:
		c.Value = v.Div(decimal.New(int64(divValue), 0)).Floor()
	case string:
		c.Value = decimal.RequireFromString(v).Div(decimal.New(int64(divValue), 0)).Floor()
	default:
		c.Value = byteData
	}
}

func (c *Compact) TypeStructString() string {
	return fmt.Sprintf("Compact<%s>", c.SubType)
}

func (c *Compact) Encode(value interface{}) string {
	var compactValue decimal.Decimal
	switch v := value.(type) {
	case uint64:
		compactValue = decimal.New(int64(v), 0)
	case decimal.Decimal:
		compactValue = v
	case int64:
		compactValue = decimal.New(v, 0)
	case int:
		compactValue = decimal.New(int64(v), 0)
	case float64:
		compactValue = decimal.NewFromFloat(v)
	case string:
		compactValue = decimal.RequireFromString(v)
	}
	bs := make([]byte, 4)
	if compactValue.IntPart() <= 63 {
		binary.LittleEndian.PutUint32(bs, uint32(compactValue.IntPart()<<2))
		c.Data.Data = bs[0:1]
	} else if compactValue.IntPart() <= 16383 {
		binary.LittleEndian.PutUint32(bs, uint32(compactValue.IntPart()<<2)|1)
		c.Data.Data = bs[0:2]
	} else if compactValue.IntPart() <= 1073741823 {
		binary.LittleEndian.PutUint32(bs, uint32(compactValue.IntPart()<<2)|2)
		c.Data.Data = bs
	} else {
		v := compactValue.BigInt()
		numBytes := len(v.Bytes())
		if numBytes > 255 || uint8(numBytes-4) > 63 {
			return ""
		}
		buf := v.Bytes()
		Reverse(buf)
		c.Data.Data = append([]byte{uint8(numBytes-4)<<2 + 3}, buf...)
	}
	return utiles.BytesToHex(c.Data.Data)
}

func Reverse(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

type CompactU32 struct {
	Compact
}

func (c *CompactU32) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	c.TypeString = "Compact<u32>"
	c.ScaleDecoder.Init(data, option)
}

func (c *CompactU32) Process() {
	c.ProcessCompactBytes()
	buf := &bytes.Buffer{}
	var reader io.Reader
	reader = buf
	_, _ = buf.Write(c.CompactBytes)
	b := make([]byte, 8)
	_, _ = reader.Read(b)
	c.Value = int(binary.LittleEndian.Uint64(b))
	if c.CompactLength <= 4 {
		c.Value = int(c.Value.(int)) / 4
	}

}

func (c *CompactU32) TypeStructString() string {
	return c.TypeString
}

func (c *CompactU32) Encode(value interface{}) string {
	var i int
	switch v := value.(type) {
	case int:
		i = v
	case decimal.Decimal:
		i = int(v.IntPart())
	case uint32:
		i = int(v)
	case int64:
		i = int(v)
	case uint16:
		i = int(v)
	case float64:
		i = int(v)
	}
	if i <= 63 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(i<<2))
		c.Data.Data = bs[0:1]
	} else if i <= 16383 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(i<<2)|1)
		c.Data.Data = bs[0:2]
	} else if i <= 1073741823 {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(i<<2)|2)
		c.Data.Data = bs
	}
	return utiles.BytesToHex(c.Data.Data)
}

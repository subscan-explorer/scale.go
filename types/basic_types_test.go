package types

import (
	"testing"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/assert"
)

func TestBytesAndHexBytes(t *testing.T) {
	raw := "1054657374"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, "Test", m.ProcessAndUpdateData("Bytes").(string))
	assert.Equal(t, raw, Encode("Bytes", "Test"))

	hexRaw := "080102"
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(hexRaw)}, nil)
	assert.Equal(t, "0x0102", m.ProcessAndUpdateData("HexBytes").(string))
	assert.Equal(t, hexRaw, Encode("HexBytes", "0x0102"))
}

func TestOptionBool(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("01")}, &ScaleDecoderOption{SubType: "bool"})
	assert.Equal(t, true, m.ProcessAndUpdateData("Option<bool>").(bool))
	assert.Equal(t, "01", Encode("Option<bool>", true))
	assert.Equal(t, "02", Encode("Option<bool>", false))
	assert.Equal(t, "00", Encode("Option<bool>", nil))
}

func TestNull(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: []byte{}}, nil)
	assert.Nil(t, m.ProcessAndUpdateData("Null"))
	assert.Equal(t, "", Encode("Null", nil))
}

func TestHashTypes(t *testing.T) {
	h256Raw := "1111111111111111111111111111111111111111111111111111111111111111"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(h256Raw)}, nil)
	assert.Equal(t, h256Raw, utiles.TrimHex(m.ProcessAndUpdateData("H256").(string)))
	assert.Equal(t, h256Raw, Encode("H256", "0x"+h256Raw))

	h512Raw := "2222222222222222222222222222222222222222222222222222222222222222" +
		"2222222222222222222222222222222222222222222222222222222222222222"
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(h512Raw)}, nil)
	assert.Equal(t, h512Raw, utiles.TrimHex(m.ProcessAndUpdateData("H512").(string)))
	assert.Equal(t, h512Raw, Encode("H512", "0x"+h512Raw))
}

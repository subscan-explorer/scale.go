package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/shopspring/decimal"
)

type IntFixed struct {
	ScaleDecoder
	FixedLength int
}

func (f *IntFixed) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	if option != nil && option.FixedLength != 0 {
		f.FixedLength = option.FixedLength
	}
	f.ScaleDecoder.Init(data, option)
}

func (f *IntFixed) Process() {
	value := utiles.U256(utiles.BytesToHex(utiles.ReverseBytes(f.NextBytes(f.FixedLength))))
	var i, e, b = big.NewInt(2), big.NewInt(int64(f.FixedLength*8) - 1), big.NewInt(int64(f.FixedLength * 8))
	unsignedMid := big.NewInt(2).Exp(i, e, nil)
	ceiling := big.NewInt(2).Exp(i, b, nil)
	if value.Cmp(unsignedMid) > 0 {
		f.Value = value.Sub(value, ceiling)
		return
	}
	f.Value = value
}

func (f *IntFixed) Encode(value interface{}) string {
	if f.FixedLength < 16 {
		var buffer = &bytes.Buffer{}
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			panic(fmt.Errorf("fixed write raise err %s", err.Error()))
		}
		c := make([]byte, f.FixedLength)
		_, _ = buffer.Read(c)
		return utiles.BytesToHex(c)
	}
	var valueBigInt *big.Int
	switch v := value.(type) {
	case float64:
		valueBigInt = decimal.NewFromFloat(v).BigInt()
	case decimal.Decimal:
		valueBigInt = v.BigInt()
	case *big.Int:
		valueBigInt = v
	case string:
		valueBigInt = decimal.RequireFromString(v).BigInt()
	default:
		panic("invalid intFixed value type")
	}
	if c, err := BigIntToIntBytes(valueBigInt, f.FixedLength); err != nil {
		panic(err)
	} else {
		return utiles.BytesToHex(c)
	}
}

func (f *IntFixed) TypeStructString() string {
	return fmt.Sprintf("Int<%d>", f.FixedLength*8)
}

// BigIntToIntBytes ref https://github.com/centrifuge/go-substrate-rpc-client/blob/master/types/int.go#L218
func BigIntToIntBytes(i *big.Int, bytelen int) ([]byte, error) {
	res := make([]byte, bytelen)

	maxNeg := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(bytelen*8-1)), nil)
	maxPos := big.NewInt(0).Sub(maxNeg, big.NewInt(1))

	if i.Sign() >= 0 {
		if i.CmpAbs(maxPos) > 0 {
			return nil, fmt.Errorf("cannot encode big.Int to []byte: given big.Int exceeds highest positive number "+
				"%v for an int with %v bits", maxPos, bytelen*8)
		}

		bs := i.Bytes()
		copy(res[len(res)-len(bs):], bs)
		return res, nil
	}

	// negative, two's complement

	if i.CmpAbs(maxNeg) > 0 {
		return nil, fmt.Errorf("cannot encode big.Int to []byte: given big.Int exceeds highest negative number -"+
			"%v for an int with %v bits", maxNeg, bytelen*8)
	}

	i = big.NewInt(0).Add(i, big.NewInt(1))
	bs := i.Bytes()
	copy(res[len(res)-len(bs):], bs)

	// apply not to every byte
	for j, b := range res {
		res[j] = ^b
	}

	return res, nil
}

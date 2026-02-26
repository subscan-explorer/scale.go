package types

import (
	"testing"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/shopspring/decimal"
)

func BenchmarkDecodeU32(b *testing.B) {
	data := utiles.HexToBytes("64000000")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := ScaleDecoder{}
		decoder.Init(scaleBytes.ScaleBytes{Data: data}, nil)
		_ = decoder.ProcessAndUpdateData("U32")
	}
}

func BenchmarkDecodeRegistrationBalanceOf(b *testing.B) {
	data := utiles.HexToBytes("04010000000200a0724e180900000000000000000000000d505552455354414b452d30310e507572655374616b65204c74641b68747470733a2f2f7777772e707572657374616b652e636f6d2f000000000d40707572657374616b65636f")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := ScaleDecoder{}
		decoder.Init(scaleBytes.ScaleBytes{Data: data}, nil)
		_ = decoder.ProcessAndUpdateData("Registration<BalanceOf>")
	}
}

func BenchmarkEncodeCompactBalance(b *testing.B) {
	value := decimal.NewFromInt32(750000000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Encode("Compact<Balance>", value)
	}
}

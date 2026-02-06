package types

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/itering/scale.go/source"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	raw := "1054657374"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, "Test", m.ProcessAndUpdateData("String").(string))
	assert.Equal(t, raw, Encode("String", "Test"))
}

func TestCompactU64(t *testing.T) {
	raw := "10"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.EqualValues(t, 4, m.ProcessAndUpdateData("Compact<U64>").(uint64))
}

func TestU32(t *testing.T) {
	raw := "64000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.EqualValues(t, 100, m.ProcessAndUpdateData("U32"))
	assert.Equal(t, raw, Encode("U32", 100))
}

func TestU16(t *testing.T) {
	raw := "0300"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.EqualValues(t, 3, m.ProcessAndUpdateData("u16"))
	assert.Equal(t, raw, Encode("U16", 3))
}

func TestRawBabePreDigest(t *testing.T) {
	raw := "02" + // enum
		"02000000" + // u32
		"8b86750900000000" + // u64
		"00000000" // u32
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, Encode("RawBabePreDigest", m.ProcessAndUpdateData("RawBabePreDigest")))
}

func TestRawBabePreDigestVRF(t *testing.T) {
	raw := "030000000099decc0f0000000040a523a6fdd15ef7ffb2956689b828185b4d60cfac789f64d1b6f26257ebbe543349f8ceae602875c705a59b156af586c7cf907df5c8d5b541fa755638e32b07b02bfb5e7549fb88aa1f32da93519c67275e999da1cd58ec168c80b3"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, Encode("RawBabePreDigest", m.ProcessAndUpdateData("RawBabePreDigest")))
}

func TestSet_Process(t *testing.T) {
	RegCustomTypes(map[string]source.TypeStruct{
		"CustomSet": {
			Type:      "set",
			BitLength: 64,
			ValueList: []string{"Value1", "Value2", "Value3", "Value4", "Value5"},
		},
	})
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x03000000")}, nil)
	r := m.ProcessAndUpdateData("CustomSet")
	if strings.Join(r.([]string), "") != "Value1Value2" {
		t.Errorf("Test TestSet_Process Process fail, decode return %v", r.([]string))
	}
}

func Test_ComplexEnum(t *testing.T) {

	RegCustomTypes(map[string]source.TypeStruct{
		"CustomEnum": {
			Type:        "enum",
			TypeMapping: [][]string{{"a", "[[\"col1\", \"u32\"], [\"col2\", \"u32\"]]"}},
		},
	})
	assert.EqualValues(t, "000100000002000000", Encode("CustomEnum", map[string]interface{}{"a": map[string]interface{}{"col1": 1, "col2": 2}}))

}

// 0x025ed0b2 Compact<Balance>
func TestCompactBalance(t *testing.T) {
	raw := "0x025ed0b2"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, decimal.NewFromInt32(750000000), m.ProcessAndUpdateData("Compact<Balance>"))
	assert.Equal(t, "04", Encode("Compact<Balance>", 1))
	assert.Equal(t, "025ed0b2", Encode("Compact<Balance>", decimal.NewFromInt32(750000000)))
	assert.Equal(t, "025ed0b2", Encode("Compact<Balance>", decimal.NewFromInt32(750000000)))
	assert.Equal(t, "0f0000c16ff28623", Encode("Compact<Balance>", decimal.RequireFromString("10000000000000000")))
	assert.Equal(t, "13000064a7b3b6e00d", Encode("Compact<Balance>", decimal.RequireFromString("1000000000000000000")))
}

// 0xe52d2254c67c430a0000000000000000 Balance
func TestBalance(t *testing.T) {
	raw := "e52d2254c67c430a0000000000000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, Encode("Balance", m.ProcessAndUpdateData("Balance")))
}

func TestRegistration(t *testing.T) {
	raw := "04010000000200a0724e180900000000000000000000000d505552455354414b452d30310e507572655374616b65204c74641b68747470733a2f2f7777772e707572657374616b652e636f6d2f000000000d40707572657374616b65636f"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, Encode("Registration<BalanceOf>", m.ProcessAndUpdateData("Registration<BalanceOf>")))
}

func TestInt(t *testing.T) {
	raw := "2efb"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, Encode("i16", m.ProcessAndUpdateData("i16").(*big.Int).Int64()))
}

func TestBoolArray(t *testing.T) {
	raw := "00000100"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.EqualValues(t, []interface{}{false, false, true, false}, m.ProcessAndUpdateData("[bool; 4]"))
}

func TestReferendumInfo(t *testing.T) {
	raw := "0x00004e0c00295ce46278975a53b855188482af699f7726fbbeac89cf16a1741c4698dcdbc90080970600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	r := m.ProcessAndUpdateData("ReferendumInfo<BlockNumber, Hash, BalanceOf>")
	c := map[string]interface{}{
		"Ongoing": map[string]interface{}{
			"delay":        432000,
			"end":          806400,
			"proposalHash": "0x295ce46278975a53b855188482af699f7726fbbeac89cf16a1741c4698dcdbc9",
			"tally":        map[string]interface{}{"ayes": "0", "nays": "0", "turnout": "0"}, "threshold": "SuperMajorityApprove",
		}}
	assert.Equal(t, utiles.ToString(c), utiles.ToString(r))
}

func TestEthereumAccountId(t *testing.T) {
	raw := "0x4119b2e6c3cb618f4f0B93ac77f9Beec7ff02887"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	r := m.ProcessAndUpdateData("EthereumAccountId")
	if r.(string) != "0x4119b2e6c3Cb618F4f0B93ac77f9BeeC7FF02887" {
		t.Errorf("Test TestEthereumAccountId Process fail, decode return %v", r)
	}
}

func TestRegistrarInfo(t *testing.T) {
	raw := "0x08014c4bf7f93d0a5ed801ef778f8e7ef58201bdd7e33e167faf42a01d439283cb430000000000000000000000000000000000000000000000000112ccb53338ac0da571d3697548346fb5f0b637ac9412f8abbf6d13588be7563200d8c3795800000000000000000000000000000000000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	m.ProcessAndUpdateData("Vec<Option<RegistrarInfo<BalanceOf, AccountId>>>")
}

func TestRewardDestinationLatest(t *testing.T) {
	raw := "0x03f8764d575b96b30e095a201d90b6ddaf944d042811846f7a3fe5ffda2a01c045"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	m.ProcessAndUpdateData("RewardDestination")
}

func TestGenericLookupSource(t *testing.T) {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(256))
	c := []byte{255, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	raw := utiles.BytesToHex(c)
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	r := m.ProcessAndUpdateData("GenericLookupSource")
	if r.(string) != "0102030405060708010203040506070801020304050607080102030405060708" {
		t.Errorf("Test TestGenericLookupSource Process fail, decode return %d", r)
	}
	m.Init(scaleBytes.ScaleBytes{Data: []byte{0xfc, 0, 1}}, nil)
	m.ProcessAndUpdateData("GenericLookupSource")
}

func TestBTreeMap(t *testing.T) {
	raw := "0x041c62617a7a696e6745000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	r := m.ProcessAndUpdateData("BTreeMap<Text,u32>")
	if utiles.ToString(r) != `[{"bazzing":69}]` {
		t.Errorf("Test TestBTreeMap Process fail, decode return %v", utiles.ToString(r))
	}
}

func TestClikeEnum(t *testing.T) {
	raw := "0x45"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	RegCustomTypes(source.LoadTypeRegistry([]byte(`{"t": {"type": "enum","type_mapping": [["A","42"],["B","69"],["C","255"]]}}`)))
	r := m.ProcessAndUpdateData("t")
	if utiles.ToString(r) != `B` {
		t.Errorf("Test TestClikeEnum Process fail, decode return %v", utiles.ToString(r))
	}
}

func TestNamespaceInt(t *testing.T) {
	raw := "0x2efb"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	r := m.ProcessAndUpdateData("Eth::i16")
	if r.(*big.Int).String() != `-1234` {
		t.Errorf("Test TestNamespaceInt Process fail, decode return %v", utiles.ToString(r))
	}
}

func TestBoundedVec(t *testing.T) {
	raw := "0x08014c4bf7f93d0a5ed801ef778f8e7ef58201bdd7e33e167faf42a01d439283cb430000000000000000000000000000000000000000000000000112ccb53338ac0da571d3697548346fb5f0b637ac9412f8abbf6d13588be7563200d8c3795800000000000000000000000000000000000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	m.ProcessAndUpdateData("BoundedVec<Option<RegistrarInfo<BalanceOf, AccountId>>,5>")
}

func TestBTreeSet(t *testing.T) {
	raw := "0x1002000000180000001e00000050000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	m.ProcessAndUpdateData("BTreeSet<U32>")
}

func TestModuleTypeOverride(t *testing.T) {
	raw := "0x21033e010000830100004403000025010000ec020000bb010000050100002a010000b8000000680300008f010000d702000040010000f00000006b02000014000000bf010000460100006200000044000000b40200004603000022000000760000004303000021030000a3000000cc000000030300002e030000ac020000c5020000f701000029010000ab010000280300001b00000091000000e7010000f8000000930200005e010000f4020000240000004b00000043020000c70100007e020000da0200000d0100000d0200008b0100000d000000d10200007d030000ef0000005b0200005c0100008a020000e3000000b1010000ad000000210200003d0300007b020000e9010000b8020000c30000008b0200003f000000330000005a03000037010000a20200004901000096020000340100007a0000004e0100004a020000d10000003f0200009e0200002e020000e0010000c101000087000000110100009e000000bd000000dc000000a20100004d000000b102000053010000040300006b01000027020000b5010000fd0000001c0000001c020000da00000048010000300300002700000080000000890100005000000017030000bd010000e20100006b03000009030000bb0000007e00000020030000e0020000ee020000730000008f0200000a000000570100000f0100001a000000e801000074020000be01000081010000a6020000f201000058010000570300002c000000b0000000a6010000420000007d0200006802000055020000010300009f020000a901000066000000b6020000a900000077000000c40000002f000000b10000001b010000c60100006c030000d0010000300200003c000000cd020000000100006f01000063020000d60000005e030000b90100005401000095010000010000000b000000140300002d02000071010000fc0200005303000005020000350300005d0300009c000000c902000064030000ea0100006202000010010000090100004e00000018020000dc02000023010000250000004f0100003b010000fb000000f501000020020000730300009202000059030000c1000000fb020000fc010000e20200004c020000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, &ScaleDecoderOption{Module: "parasShared"})
	m.ProcessAndUpdateData("Vec<ValidatorIndex>")
}

func TestAccountInfo(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(`0x0100000001000000`)}, nil)
	m.ProcessAndUpdateData("AccountInfo<Index, AccountData>")
}

func TestWeakBoundedVec(t *testing.T) {
	raw := "0x047374616b696e672063fae72d71290000000000000000000002"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	m.ProcessAndUpdateData("WeakBoundedVec<BalanceLock<Balance>, MaxLocks>")
}

func TestNestFixedArray(t *testing.T) {
	raw := "0x010101010101010101"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(
		t,
		[]interface{}{"0x010101", "0x010101", "0x010101"},
		m.ProcessAndUpdateData("[[u8; 3]; 3]"),
	)
}

func TestTypeIsStruct(t *testing.T) {
	c := "[[\"assets\", \"MultiAssetFilterV1\"], [\"maxAssets\", \"u32\"], [\"beneficiary\", \"MultiLocationV1\"]]"
	var typeMap [][]string
	if c[0:2] == "[[" && json.Unmarshal([]byte(c), &typeMap) == nil && len(typeMap) > 0 && len(typeMap[0]) == 2 {
		result := make(map[string]interface{})
		for _, v := range typeMap {
			result[v[0]] = v[1]
		}
		assert.Equal(t, result, map[string]interface{}{"assets": "MultiAssetFilterV1", "beneficiary": "MultiLocationV1", "maxAssets": "u32"})
	}
}

func TestCompactU32_Encode(t *testing.T) {
	compactU32 := CompactU32{}
	compactU32.Encode(100)
	assert.Equal(t, "0x9101", compactU32.Data.String())
}

func TestTupleDisassemble(t *testing.T) {
	assert.Equal(t, []string{"U32"}, TupleDisassemble("U32"))
	assert.Equal(t, []string{"U32", "U32"}, TupleDisassemble("(U32,U32)"))
	assert.Equal(t, []string{"U32", "(U32,U64)"}, TupleDisassemble("(U32,(U32,U64))"))
	assert.Equal(t, []string{"(U32,U16)", "(U32,U64)"}, TupleDisassemble("((U32,U16),(U32,U64))"))
}

func TestWWrapperOpaque(t *testing.T) {
	raw := []byte{4 << 2, 135, 214, 18, 0}
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: raw}, nil)
	value := m.ProcessAndUpdateData("WrapperOpaque<u32>")
	assert.Equal(t, 1234567, int(value.(uint32)))
}

func TestRange(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: []byte{1, 0, 0, 0, 2, 0, 0, 0}}, nil)
	assert.Equal(t, map[string]interface{}{"start": uint32(1), "end": uint32(2)}, m.ProcessAndUpdateData("Range<u32>"))
}

func TestEncode(t *testing.T) {
	assert.Equal(t, "00000000", Encode("U32", 0))
	assert.Equal(t, "010000000200000003000000040000000500000006000000", Encode("[U32; 6]", []uint32{1, 2, 3, 4, 5, 6}))
	assert.Equal(t, "01000000", Encode("U32", decimal.NewFromInt32(1)))
	assert.Equal(t, "0100", Encode("U16", decimal.NewFromInt32(1)))
	assert.Equal(t, "0d00", Encode("U16", 13))
	assert.Equal(t, "0d00", Encode("U16", int64(13)))
	assert.Equal(t, "00000000000000000000000000000000", Encode("U128", decimal.Zero))
	assert.Equal(t, "87d61200000000000000000000000000", Encode("U128", decimal.NewFromInt32(1234567)))
	assert.Equal(t, "070736c8230a", Encode("compact<u128>", decimal.New(43549996551, 0)))
	assert.Equal(t, "87d61200000000000000000000000000", Encode("U128", decimal.NewFromInt32(1234567)))
	assert.Equal(t, "47a952404f2b1568e6881efeb58c8918", Encode("U128", decimal.RequireFromString("32615670524745285411807346420584982855")))
	assert.Equal(t, "9a5b8a1b7bca89cdb3931d8ee71aa468081d971c", Encode("H160", "0x9a5b8a1B7Bca89CDB3931D8eE71AA468081D971c"))
}

func TestSubstrateFixedU64(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("001e85eb01000000")}, nil)
	assert.Equal(t, "1.919921875", m.ProcessAndUpdateData("SubstrateFixedU64").(decimal.Decimal).String())
}

func TestFloat(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x0000000000000080")}, nil)
	assert.Equal(t, float64(0), m.ProcessAndUpdateData("Float64"))
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x00000080")}, nil)
	assert.Equal(t, float32(0), m.ProcessAndUpdateData("Float32"))
}

func TestDataRawDecode(t *testing.T) {
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x00a41a130d84010000000000000000000000076f6e64696e33000000136f6e64696e37373740676d61696c2e636f6d000000")}, nil)
	value := m.ProcessAndUpdateData("Registration").(map[string]interface{})
	assert.Equal(t, map[string]interface{}{"Raw18": "ondin777@gmail.com"}, value["info"].(map[string]interface{})["email"])
}

func TestEmptyVec(t *testing.T) {
	raw := "00"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.EqualValues(t, []interface{}(nil), m.ProcessAndUpdateData("Vec<u32>"))
	assert.Equal(t, raw, Encode("Vec<u32>", nil))
}

func TestBitVec(t *testing.T) {
	m := ScaleDecoder{}
	r := []string{"0b00010111", "0b01111011", "0b00110011"}
	for i := range []string{"0x0817", "0x087b", "0x0c33"} {
		m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(utiles.AddHex(Encode("BitVec", r[i])))}, nil)
		assert.EqualValues(t, r[i], m.ProcessAndUpdateData("BitVec"))
	}
}

func TestFixedArray(t *testing.T) {
	ts := []string{"[u8; 16]", "[u8; 16]", "[u8;16]"}
	for i, v := range []interface{}{[]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}, []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}, "0x02020202020202020202020202020202"} {
		assert.Equal(t, Encode(ts[i], v), "02020202020202020202020202020202", "TestFixedArray Encode fail %s", ts[i])
	}
}

func TestVecEncodeSliceTypes(t *testing.T) {
	assert.Equal(t, "080100000002000000", Encode("Vec<u32>", []uint32{1, 2}))
	assert.Equal(t, "080100000002000000", Encode("Vec<u32>", []int{1, 2}))
}

func TestU256(t *testing.T) {
	raw := "0x1001000000000000000000000000000000000000000000000000000000000000"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, raw, utiles.AddHex(Encode("U256", m.ProcessAndUpdateData("U256").(decimal.Decimal))))
	assert.Equal(t, raw, utiles.AddHex(Encode("U256", "0x1001000000000000000000000000000000000000000000000000000000000000")))
	assert.Equal(t, raw, utiles.AddHex(Encode("U256", utiles.HexToBytes(raw))))
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x00b5070000000000000000000000000000000000000000000000000000000000")}, nil)
	assert.Equal(t, int64(505088), m.ProcessAndUpdateData("U256").(decimal.Decimal).IntPart())
}

func TestXcmV2ResultType(t *testing.T) {
	raw := "0204031501020042f2f9dc"
	m := ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, nil)
	assert.Equal(t, `{"V2":[{"QueryResponse":{"max_weight":0,"query_id":69,"response":{"ExecutionResult":{"Ok":3707368002}}}}]}`, utiles.ToString(m.ProcessAndUpdateData("VersionedXcm").(map[string]interface{})))
}

func Test_ConcurrencyDecode(t *testing.T) {
	wg := sync.WaitGroup{}
	goRoutines := 1000
	for i := 0; i < goRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := ScaleDecoder{}
			m.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("0x2d63bf0800000000")}, nil)
			if r := m.ProcessAndUpdateData("RawAuraPreDigest").(map[string]interface{}); r["slotNumber"] != uint64(146760493) {
				t.Errorf("Test Test_ConcurrencyDecode Process fail, decode return %v", r)
			}
		}()
	}
	wg.Wait()
}

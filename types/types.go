package types

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/huandu/xstrings"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/itering/scale.go/utiles/crypto/ethereum"
	"github.com/itering/scale.go/utiles/uint128"
	"github.com/shopspring/decimal"
)

type H160 struct{ ScaleDecoder }

type Other struct{ HexBytes }

type ChangesTrieRoot struct{ HexBytes }

type BTreeSet struct{ Vec }

type RuntimeEnvironmentUpdated struct{ Null }

type SlotNumber struct{ U64 }

type Null struct{ ScaleDecoder }

func (*Null) TypeStructString() string {
	return "NULL"
}

type Empty struct{ ScaleDecoder }

func (e *Empty) Process() {
	e.Value = "NULL"
}

func (e *Empty) Encode(interface{}) string { return "" }

func (h *H160) Process() {
	h.Value = utiles.AddHex(utiles.BytesToHex(h.NextBytes(20)))
}

func (h *H160) Encode(value string) string {
	return utiles.AddHex(strings.ToLower(value))
}

func (h *H160) TypeStructString() string {
	return "H160"
}

type H256 struct {
	ScaleDecoder
}

func (h *H256) Process() {
	h.Value = utiles.AddHex(utiles.BytesToHex(h.NextBytes(32)))
}

func (h *H256) Encode(value string) string {
	return utiles.AddHex(strings.ToLower(value))
}

func (h *H256) TypeStructString() string {
	return "H256"
}

type H512 struct {
	ScaleDecoder
}

func (h *H512) Process() {
	h.Value = utiles.AddHex(utiles.BytesToHex(h.NextBytes(64)))
}

func (h *H512) Encode(value string) string {
	return utiles.AddHex(strings.ToLower(value))
}

func (h *H512) TypeStructString() string {
	return "H512"
}

type Era struct {
	ScaleDecoder
}

func (e *Era) Process() {
	optionHex := utiles.BytesToHex(e.NextBytes(1))
	if optionHex == "00" {
		e.Value = optionHex
	} else {
		e.Value = optionHex + utiles.BytesToHex(e.NextBytes(1))
	}
}

func (e *Era) Encode(era string) string {
	return era
}

type EraExtrinsic struct{ Era }

type CompactMoment struct {
	CompactU32
}

func (m *CompactMoment) Process() {
	m.CompactU32.Process()
	if m.Value.(int) > 10000000000 {
		m.Value = m.Value.(int) / 1000
	}
}

type Moment struct {
	U64
}

func (m *Moment) Process() {
	m.U64.Process()
	if m.Value.(uint64) > 10000000000 {
		m.Value = m.Value.(uint64) / 1000
	}
}

type BlockNumber struct {
	U32
}

type Address struct {
	ScaleDecoder
	AccountLength string `json:"account_length"`
}

func (a *Address) Process() {
	AccountLength := a.NextBytes(1)
	a.AccountLength = utiles.BytesToHex(AccountLength)
	if a.AccountLength == "ff" {
		a.Value = utiles.BytesToHex(a.NextBytes(32))
		return
	}
	a.Value = utiles.BytesToHex(append(AccountLength, a.NextBytes(31)...))
}

type GenericAddress struct {
	ScaleDecoder
	AccountLength string `json:"account_length"`
}

func (a *GenericAddress) Process() {
	AccountLength := a.NextBytes(1)
	a.AccountLength = utiles.BytesToHex(AccountLength)
	if a.AccountLength == "ff" {
		a.Value = utiles.BytesToHex(a.NextBytes(32))
		return
	}
	switch a.AccountLength {
	case "fc":
		a.NextBytes(2)
	case "fd":
		a.NextBytes(4)
	case "fe":
		a.NextBytes(8)
	default:
		a.Value = utiles.BytesToHex(append(AccountLength, a.NextBytes(31)...))
	}
}

type Signature struct {
	H512
}

type AccountId struct {
	ScaleDecoder
}

func (s *AccountId) Process() {
	s.Value = utiles.AddHex(xstrings.RightJustify(utiles.BytesToHex(s.NextBytes(32)), 64, "0"))
}

func (s *AccountId) Encode(value string) string {
	return value
}

func (s *AccountId) TypeStructString() string {
	return "AccountId"
}

type Balance struct {
	U128
}

func (b *Balance) Process() {
	buf := &bytes.Buffer{}
	var reader io.Reader
	reader = buf
	_, _ = buf.Write(b.NextBytes(16))
	c := make([]byte, 16)
	_, _ = reader.Read(c)
	if utiles.BytesToHex(c) == "ffffffffffffffffffffffffffffffff" {
		b.Value = decimal.NewFromInt32(-1)
		return
	}
	b.Value = decimal.NewFromBigInt(uint128.FromBytes(c).Big(), 0)
}

type LogDigest struct{ Enum }

func (l *LogDigest) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	l.ValueList = []string{"Other", "AuthoritiesChange", "ChangesTrieRoot", "SealV0", "Consensus", "Seal", "PreRuntime", "ChangesTrieSignal", "RuntimeEnvironmentUpdated"}
	l.Enum.Init(data, option)
}

func (l *LogDigest) Process() {
	index := utiles.BytesToHex(l.NextBytes(1))
	if utiles.U256(index) != nil {
		l.Index = int(utiles.U256(index).Uint64())
	}
	indexType := l.ValueList[l.Index]
	if indexType == "" {
		panic(fmt.Sprintf("LogDigest index %d not in list", l.Index))
	}
	l.Value = map[string]interface{}{
		"type":  indexType,
		"value": l.ProcessAndUpdateData(indexType),
	}
}

type AuthoritiesChange struct{ Vec }

func (l *AuthoritiesChange) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	option.SubType = "AccountId"
	l.Vec.Init(data, option)
}

type SealV0 struct{ Struct }

func (s *SealV0) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"slot", "signature"}, Types: []string{"u64", "Signature"}}
	s.Struct.Init(data, option)
}

type Consensus struct{ Struct }

func (s *Consensus) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"engine", "data"}, Types: []string{"u32", "Vec<u8>"}}
	s.Struct.Init(data, option)
}

type Seal struct{ Struct }

func (s *Seal) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"engine", "data"}, Types: []string{"u32", "HexBytes"}}
	s.Struct.Init(data, option)
}

type PreRuntime struct{ Struct }

func (s *PreRuntime) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"engine", "data"}, Types: []string{"u32", "HexBytes"}}
	s.Struct.Init(data, option)
}

type Exposure struct{ Struct }

func (s *Exposure) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"total", "own", "others"}, Types: []string{"Compact<Balance>", "Compact<Balance>", "Vec<IndividualExposure<AccountId, Balance>>"}}
	s.Struct.Init(data, option)
}

type IndividualExposure struct{ Struct }

func (s *IndividualExposure) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"who", "value"}, Types: []string{"AccountId", "Compact<Balance>"}}
	s.Struct.Init(data, option)
}

type RawAuraPreDigest struct{ Struct }

func (s *RawAuraPreDigest) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	s.Struct.TypeMapping = &TypeMapping{Names: []string{"slotNumber"}, Types: []string{"u64"}}
	s.Struct.Init(data, option)
}

type RawBabePreDigestPrimary struct{ Struct }

func (r *RawBabePreDigestPrimary) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	r.Struct.TypeMapping = &TypeMapping{
		Names: []string{"authorityIndex", "slotNumber", "weight", "vrfOutput", "vrfProof"},
		Types: []string{"u32", "SlotNumber", "BabeBlockWeight", "H256", "H256"},
	}
	r.Struct.Init(data, option)
}

type RawBabePreDigestSecondary struct{ Struct }

func (r *RawBabePreDigestSecondary) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	r.Struct.TypeMapping = &TypeMapping{Names: []string{"authorityIndex", "slotNumber", "weight"}, Types: []string{"u32", "SlotNumber", "BabeBlockWeight"}}
	r.Struct.Init(data, option)
}

type RawBabePreDigestSecondaryVRF struct{ Struct }

func (r *RawBabePreDigestSecondaryVRF) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	r.Struct.TypeMapping = &TypeMapping{Names: []string{"authorityIndex", "slotNumber", "vrfOutput", "vrfProof"}, Types: []string{"u32", "SlotNumber", "VrfData", "VrfProof"}}
	r.Struct.Init(data, option)
}

type LockIdentifier struct {
	ScaleDecoder
}

func (l *LockIdentifier) Process() {
	l.Value = utiles.AddHex(utiles.BytesToHex(l.NextBytes(8)))
}

type EcdsaSignature struct {
	ScaleDecoder
}

func (e *EcdsaSignature) Process() {
	e.Value = utiles.AddHex(utiles.BytesToHex(e.NextBytes(65)))
}

type EthereumAddress struct {
	H160
}

func (e *EthereumAddress) Process() {
	e.Value = ethereum.Encode(utiles.BytesToHex(e.NextBytes(20)))
}

type Data struct {
	Enum
}

func (d *Data) Init(data scaleBytes.ScaleBytes, option *ScaleDecoderOption) {
	d.TypeMapping = &TypeMapping{
		Names: []string{"None", "Raw", "BlakeTwo256", "Sha256", "Keccak256", "ShaThree256"},
		Types: []string{"Null", "String", "H256", "H256", "H256", "H256"},
	}
	d.Enum.Init(data, option)
}

func (d *Data) Process() {
	c := utiles.BytesToHex(d.NextBytes(1))
	if c == "" || utiles.U256(c).Uint64() == 0 {
		d.Value = map[string]interface{}{"None": nil}
		return
	}

	d.Index = int(utiles.U256(c).Uint64())

	if d.Index >= 1 && d.Index <= 33 {
		data := d.NextBytes(d.Index - 1)
		if utf8.Valid(data) {
			d.Value = map[string]interface{}{fmt.Sprintf("Raw%d", d.Index-1): string(data)}
		} else {
			d.Value = map[string]interface{}{fmt.Sprintf("Raw%d", d.Index-1): utiles.BytesToHex(data)}
		}

	}
	if d.Index >= 34 && d.Index <= 37 {
		enumKey := d.TypeMapping.Names[d.Index-32]
		d.Value = map[string]interface{}{enumKey: d.ProcessAndUpdateData(d.TypeMapping.Types[d.Index-32])}
	}
}

func (d *Data) Encode(v map[string]interface{}) string {
	key, val, err := utiles.GetEnumValue(v)
	if err != nil {
		panic(err)
	}
	if index := utiles.SliceIndex(key, d.TypeMapping.Names); index > -1 {
		if key == "None" {
			return Encode("U8", 0)
		}
		return Encode("U8", index+32) + Encode(d.TypeMapping.Types[index], val.(string))
	}
	// raw data
	if strings.HasPrefix(key, "Raw") {
		var l int
		if _, err := fmt.Sscanf(key, "Raw%d", &l); err != nil {
			panic("invalid data Raw data key")
		} else {
			l++
			indexRaw := Encode("U8", l)
			if l == len(val.(string)) {
				return indexRaw + val.(string)
			}
			return indexRaw + utiles.BytesToHex([]byte(val.(string)))
		}
	}
	panic("invalid enum key")
}

type VoteOutcome struct{ H256 }

type OpaqueCall struct {
	Bytes
}

func (f *OpaqueCall) Process() {
	f.Bytes.Process()
	e := ScaleDecoder{}
	option := ScaleDecoderOption{Metadata: f.Metadata, Spec: f.Spec}
	e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(f.Value.(string))}, &option)
	value := e.ProcessAndUpdateData("Call")
	f.InternalCall = append(f.InternalCall, e.InternalCall...)
	f.Value = value
}

type GenericLookupSource struct {
	ScaleDecoder
}

func (g *GenericLookupSource) Process() {
	if len(g.Data.Data) == 32 {
		g.Value = utiles.BytesToHex(g.NextBytes(32))
	}
	AccountLength := g.NextBytes(1)
	accountLength := utiles.BytesToHex(AccountLength)
	if accountLength == "ff" {
		g.Value = utiles.BytesToHex(g.NextBytes(32))
		return
	}
	offset, length := func(value []byte) (int, int) {
		first := value[0]
		switch first {
		case 0xfc:
			return 1, 2
		case 0xfd:
			return 1, 4
		case 0xfe:
			return 1, 8
		}
		return 0, 1
	}(AccountLength)
	e := ScaleDecoder{}
	e.Init(scaleBytes.ScaleBytes{Data: g.Data.Data[offset : offset+length]}, nil)
	g.Value = strconv.Itoa(int(e.ProcessAndUpdateData("U32").(uint32)))
}

type Box struct{ ScaleDecoder }

func (b *Box) Process() {
	b.Value = b.ProcessAndUpdateData(b.SubType)
}

type WrapperOpaque struct{ ScaleDecoder }

func (w *WrapperOpaque) Process() {
	w.ProcessAndUpdateData("Compact<u32>")
	w.Value = w.ProcessAndUpdateData(w.SubType)
}

type Range struct{ ScaleDecoder }
type RangeInclusive struct{ Range }

func (r *Range) Process() {
	subTypeSlice := strings.Split(r.SubType, ",")
	start := r.SubType
	end := r.SubType
	if len(subTypeSlice) == 2 {
		start = subTypeSlice[0]
		end = subTypeSlice[1]
	}
	r.Value = map[string]interface{}{
		"start": r.ProcessAndUpdateData(start),
		"end":   r.ProcessAndUpdateData(end),
	}
}

type BitVec struct {
	Compact
}

func (b *BitVec) Process() {
	length := b.ProcessAndUpdateData("Compact<u32>").(int)
	u8a := b.NextBytes(int(math.Ceil(float64(length) / 8)))
	var value []string
	for _, v := range u8a {
		value = append(value, fmt.Sprintf("%08b", v))
	}
	b.Value = "0b" + strings.Join(value, "_")
}

func (b *BitVec) TypeStructString() string {
	return "BitVec"
}

func (b *BitVec) Encode(value string) string {
	value = strings.TrimPrefix(value, "0b")
	values := strings.Split(value, "_")
	var u8a []byte
	for _, v := range values {
		b, _ := strconv.ParseUint(v, 2, 8)
		u8a = append(u8a, byte(b))
	}
	return Encode("Compact<u32>", len(u8a)) + utiles.BytesToHex(u8a)
}

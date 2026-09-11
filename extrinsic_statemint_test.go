package scalecodec_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

func TestStatemintV16GeneralExtrinsics(t *testing.T) {
	raw, err := os.ReadFile("testdata/statemint_2005000_v16.scale")
	require.NoError(t, err)
	m := scalecodec.MetadataDecoder{}
	m.Init(raw)
	m.Spec = 2005000
	require.NotPanics(t, func() { require.NoError(t, m.Process()) })
	require.Equal(t, len(raw), m.Data.Offset, "consume the complete metadata response")
	require.Equal(t, 16, m.Metadata.MetadataVersion)
	require.Len(t, m.Metadata.Extrinsic.TransactionExtensionsByVersion, 2)
	require.Len(t, m.Metadata.Extrinsic.SignedExtensions, 13, "legacy signed extrinsics use version zero")
	views := 0
	for _, module := range m.Metadata.Metadata.Modules {
		for _, view := range module.ViewFunctions {
			require.Len(t, utiles.HexToBytes(view.ID), 32)
			views++
		}
	}
	require.Positive(t, views)
	fixtureBytes, err := os.ReadFile("testdata/statemint_2005000_extrinsics.json")
	require.NoError(t, err)
	var fixture struct {
		General []struct {
			Index     string          `json:"index"`
			Raw       string          `json:"raw"`
			Account   string          `json:"account"`
			Signature string          `json:"signature"`
			Nonce     int             `json:"nonce"`
			Era       string          `json:"era"`
			Calls     json.RawMessage `json:"calls"`
		} `json:"general"`
		Controls []string `json:"controls"`
	}
	require.NoError(t, json.Unmarshal(fixtureBytes, &fixture))
	require.Len(t, fixture.General, 17)
	option := types.ScaleDecoderOption{Metadata: &m.Metadata, Spec: 2005000}
	for _, f := range fixture.General {
		t.Run(f.Index, func(t *testing.T) {
			for _, prefixed := range []bool{true, false} {
				input := utiles.HexToBytes(f.Raw)
				if !prefixed {
					input = input[2:]
				} // All captured general extrinsics have a two-byte compact length.
				e := scalecodec.ExtrinsicDecoder{}
				e.Init(scaleBytes.ScaleBytes{Data: input}, &option)
				require.NotPanics(t, e.Process)
				require.Equal(t, len(input), e.Data.Offset)
				v := e.Value.(*scalecodec.GenericExtrinsic)
				require.Equal(t, "45", v.VersionInfo)
				require.Equal(t, 1, v.TransactionExtensionVersion)
				require.Equal(t, "Utility", v.CallModule)
				require.Equal(t, "force_batch", v.CallModuleFunction)
				require.Equal(t, f.Account, v.AccountId)
				require.Equal(t, f.Account, v.Signer)
				require.Equal(t, utiles.TrimHex(f.Signature), utiles.TrimHex(v.Signature))
				require.Equal(t, f.Nonce, v.Nonce)
				require.Equal(t, f.Era, v.Era)
				require.True(t, v.Tip.IsZero())
				gotCalls, err := json.Marshal(v.Params[0].Value)
				require.NoError(t, err)
				require.JSONEq(t, string(f.Calls), string(gotCalls))
				hash := blake2b.Sum256(utiles.HexToBytes(f.Raw))
				require.Equal(t, "0x"+hex.EncodeToString(hash[:]), v.ExtrinsicHash)
				encoded, err := v.Encode(&option)
				require.NoError(t, err)
				require.Equal(t, utiles.TrimHex(f.Raw), encoded, "round trip every extension, including None options")
			}
		})
	}
	for i, raw := range fixture.Controls {
		e := scalecodec.ExtrinsicDecoder{}
		e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, &option)
		require.NotPanics(t, e.Process)
		require.Equal(t, len(utiles.HexToBytes(raw)), e.Data.Offset)
		require.Equal(t, []string{"set", "remark_with_event", "transfer_keep_alive"}[i], e.Value.(*scalecodec.GenericExtrinsic).CallModuleFunction)
	}
	t.Run("disabled authorization", func(t *testing.T) {
		e := scalecodec.ExtrinsicDecoder{}
		e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(fixture.General[0].Raw)}, &option)
		e.Process()
		v := e.Value.(*scalecodec.GenericExtrinsic)
		v.SignedExtensions["VerifyMultiSignature"] = "Disabled"
		encoded, err := v.Encode(&option)
		require.NoError(t, err)
		unsigned := scalecodec.ExtrinsicDecoder{}
		unsigned.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(encoded)}, &option)
		require.NotPanics(t, unsigned.Process)
		decoded := unsigned.Value.(*scalecodec.GenericExtrinsic)
		require.Equal(t, len(utiles.HexToBytes(encoded)), unsigned.Data.Offset)
		require.Nil(t, decoded.AccountId)
		require.Nil(t, decoded.Signer)
		require.Empty(t, decoded.Signature)
		require.Equal(t, map[string]interface{}{"Disabled": "NULL"}, decoded.SignedExtensions["VerifyMultiSignature"])
		roundTrip, err := decoded.Encode(&option)
		require.NoError(t, err)
		require.Equal(t, encoded, roundTrip)
	})
}

func TestGeneralExtrinsicVersionSelection(t *testing.T) {
	metadata := types.MetadataStruct{
		MetadataVersion: 16,
		Extrinsic: &types.ExtrinsicMetadata{
			Versions: "0405",
			TransactionExtensions: []types.TransactionExtensionMetadata{
				{Identifier: "Unused", TypeString: "U32"},
				{Identifier: "CheckMortality", TypeString: "EraExtrinsic"},
				{Identifier: "CheckNonce", TypeString: "Compact<U32>"},
				{Identifier: "ChargeTransactionPayment", TypeString: "Compact<U128>"},
				{Identifier: "Optional", TypeString: "Option<U8>"},
			},
			TransactionExtensionsByVersion: []types.TransactionExtensionsByVersion{
				{Version: 0, Extensions: []int{1, 2, 3, 4}},
				{Version: 1, Extensions: []int{4, 3, 2, 1}},
			},
		},
		CallIndex: map[string]types.CallIndex{"0000": {Module: types.MetadataModules{Name: "System"}, Call: types.MetadataCalls{Name: "remark"}}},
	}
	option := types.ScaleDecoderOption{Metadata: &metadata}
	for version, raw := range []string{"204500000c14000000", "20450100140c000000"} {
		e := scalecodec.ExtrinsicDecoder{}
		e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(raw)}, &option)
		require.NotPanics(t, e.Process)
		require.Equal(t, 9, e.Data.Offset)
		v := e.Value.(*scalecodec.GenericExtrinsic)
		require.Equal(t, version, v.TransactionExtensionVersion)
		require.Equal(t, "remark", v.CallModuleFunction)
		require.Equal(t, 3, v.Nonce)
		require.Equal(t, "00", v.Era)
		require.Equal(t, "5", v.Tip.String())
		require.Nil(t, v.AccountId)
		require.Contains(t, v.SignedExtensions, "Optional")
		require.Nil(t, v.SignedExtensions["Optional"])
		encoded, err := v.Encode(&option)
		require.NoError(t, err)
		require.Equal(t, raw, encoded)
		delete(v.SignedExtensions, "Optional")
		_, err = v.Encode(&option)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing extension Optional")
	}
	metadata.Extrinsic.TransactionExtensionsByVersion = nil
	e := scalecodec.ExtrinsicDecoder{}
	e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes("1045010000")}, &option)
	require.PanicsWithError(t, "decode general extrinsic: unsupported transaction extension version 1", e.Process)
	require.Equal(t, 3, e.Data.Offset, "reject before interpreting extension bytes")
	v := scalecodec.GenericExtrinsic{VersionInfo: "45", TransactionExtensionVersion: 1}
	_, err := v.Encode(&option)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported transaction extension version 1")
}

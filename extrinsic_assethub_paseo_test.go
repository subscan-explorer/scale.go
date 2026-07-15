package scalecodec_test

import (
	"encoding/json"
	"os"
	"testing"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assethubPaseoExtrinsicFixture struct {
	Metadata  string `json:"metadata"`
	Extrinsic string `json:"extrinsic"`
}

func TestV14ExtrinsicDecoderAssethubPaseoSignedExtensions(t *testing.T) {
	raw, err := os.ReadFile("testdata/assethub_paseo_10981619.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var fixture assethubPaseoExtrinsicFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	m := scalecodec.MetadataDecoder{}
	m.Init(utiles.HexToBytes(fixture.Metadata))
	_ = m.Process()

	e := scalecodec.ExtrinsicDecoder{}
	option := types.ScaleDecoderOption{Metadata: &m.Metadata, Spec: 2004000}
	e.Init(scaleBytes.ScaleBytes{Data: utiles.HexToBytes(fixture.Extrinsic)}, &option)

	require.NotPanics(t, func() {
		e.Process()
	})

	extrinsic := e.Value.(*scalecodec.GenericExtrinsic)
	assert.Equal(t, "1f09", extrinsic.CallCode)
	assert.Equal(t, "PolkadotXcm", extrinsic.CallModule)
	assert.Equal(t, "limited_teleport_assets", extrinsic.CallModuleFunction)
	assert.Equal(t, 9560, extrinsic.Nonce)
	assert.Equal(t, "5502", extrinsic.Era)
	assert.Equal(t, "0", extrinsic.Tip.String())
	assert.Equal(t, false, extrinsic.SignedExtensions["RestrictOrigins"])
	assert.Equal(t, "Disabled", extrinsic.SignedExtensions["CheckMetadataHash"])
}

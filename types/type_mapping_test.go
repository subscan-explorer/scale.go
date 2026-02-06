package types

import (
	"testing"

	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/stretchr/testify/assert"
)

func TestGetTypeMappingConsensus(t *testing.T) {
	r := RuntimeType{}
	dec, _, err := r.GetCodec("Consensus", 0)
	assert.NoError(t, err)
	dec.Init(scaleBytes.EmptyScaleBytes(), &ScaleDecoderOption{})

	getter, ok := dec.(TypeMappingGetter)
	assert.True(t, ok)
	tm := getter.GetTypeMapping()
	if assert.NotNil(t, tm) {
		assert.Equal(t, []string{"engine", "data"}, tm.Names)
		assert.Equal(t, []string{"u32", "Vec<u8>"}, tm.Types)
	}
}

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

func TestGetCodecFixedArrayReturnsFreshInstance(t *testing.T) {
	resetCodecCache()
	r := RuntimeType{}
	first, _, err := r.GetCodec("[u16; 2]", 0)
	assert.NoError(t, err)
	second, _, err := r.GetCodec("[u16; 2]", 0)
	assert.NoError(t, err)
	assert.IsType(t, &FixedArray{}, first)
	assert.IsType(t, &FixedArray{}, second)
	assert.NotSame(t, first, second)
}

package contract

import (
	"os"
	"testing"

	"github.com/itering/scale.go/types"
	"github.com/stretchr/testify/assert"
)

func Test_AbiParse(t *testing.T) {
	c, err := os.ReadFile("metadata.json")
	assert.NoError(t, err)

	abi, err := InitAbi(c)
	assert.NoError(t, err)
	assert.Greater(t, len(abi.Types), 1)

	sc := types.ScaleDecoder{DuplicateName: make(map[string]int), RegisteredSiType: make(map[int]string)}
	abi.Register(&sc, "pre")

	assert.Equal(t, len(abi.RegisteredSiType), 8)

}

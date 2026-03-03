package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareChangesIncludeTypeStruct(t *testing.T) {
	prevPlainType := "u32"
	currentPlainType := "Vec<u8>"
	prev := MetadataTag{
		Modules: []MetadataModules{
			{
				Name: "Balances",
				Calls: []MetadataCalls{
					{
						Name: "set",
						Args: []MetadataModuleCallArgument{{Name: "value", Type: prevPlainType}},
					},
				},
				Events: []MetadataEvents{
					{
						Name: "Changed",
						Args: []string{prevPlainType},
					},
				},
				Storage: []MetadataStorage{
					{
						Name: "State",
						Type: StorageType{Origin: "PlainType", PlainType: &prevPlainType},
					},
				},
			},
		},
	}
	current := MetadataTag{
		Modules: []MetadataModules{
			{
				Name: "Balances",
				Calls: []MetadataCalls{
					{
						Name: "set",
						Args: []MetadataModuleCallArgument{{Name: "value", Type: currentPlainType}},
					},
				},
				Events: []MetadataEvents{
					{
						Name: "Changed",
						Args: []string{currentPlainType},
					},
				},
				Storage: []MetadataStorage{
					{
						Name: "State",
						Type: StorageType{Origin: "PlainType", PlainType: &currentPlainType},
					},
				},
			},
		},
	}

	result := current.Compare(&prev)
	moduleChanges := result.ModuleChanges["Balances"]

	assert.Len(t, moduleChanges.Calls.Changes, 1)
	assert.Equal(t, "set(U32)", moduleChanges.Calls.Changes[0].PrevTypeStruct)
	assert.Equal(t, "set(Bytes)", moduleChanges.Calls.Changes[0].CurrentTypeStruct)

	assert.Len(t, moduleChanges.Events.Changes, 1)
	assert.Equal(t, "Changed(U32)", moduleChanges.Events.Changes[0].PrevTypeStruct)
	assert.Equal(t, "Changed(Bytes)", moduleChanges.Events.Changes[0].CurrentTypeStruct)

	assert.Len(t, moduleChanges.Storage.Changes, 1)
	assert.Equal(t, "Balances.State: U32", moduleChanges.Storage.Changes[0].PrevTypeStruct)
	assert.Equal(t, "Balances.State: Bytes", moduleChanges.Storage.Changes[0].CurrentTypeStruct)
}

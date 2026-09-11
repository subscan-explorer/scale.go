package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtrinsicMetadataExtensionsForVersion(t *testing.T) {
	legacy := &ExtrinsicMetadata{SignedExtensions: []SignedExtensions{{Identifier: "Legacy", TypeString: "U8"}}}
	got, err := legacy.ExtensionsForVersion(0)
	require.NoError(t, err)
	require.Equal(t, legacy.SignedExtensions, got)
	_, err = legacy.ExtensionsForVersion(1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires versioned metadata")
	metadata := &ExtrinsicMetadata{
		Versions: "0405",
		TransactionExtensions: []TransactionExtensionMetadata{
			{Identifier: "First", Type: 1, TypeString: "U8", Implicit: 9},
			{Identifier: "Second", Type: 2, TypeString: "U16", Implicit: 8},
		},
		TransactionExtensionsByVersion: []TransactionExtensionsByVersion{
			{Version: 0, Extensions: []int{0}},
			{Version: 1, Extensions: []int{1, 0}},
			{Version: 2, Extensions: []int{}},
		},
	}
	got, err = metadata.ExtensionsForVersion(1)
	require.NoError(t, err)
	require.Equal(t, []SignedExtensions{
		{Identifier: "Second", Type: 2, TypeString: "U16", AdditionalSigned: 8},
		{Identifier: "First", Type: 1, TypeString: "U8", AdditionalSigned: 9},
	}, got)
	got, err = metadata.ExtensionsForVersion(2)
	require.NoError(t, err)
	require.Empty(t, got)
	for _, version := range []int{-1, 3, 256} {
		_, err = metadata.ExtensionsForVersion(version)
		require.Error(t, err)
	}
	for _, index := range []int{-1, 2} {
		invalid := *metadata
		invalid.TransactionExtensionsByVersion = []TransactionExtensionsByVersion{{Version: 1, Extensions: []int{index}}}
		_, err = invalid.ExtensionsForVersion(1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid extension index")
	}
	unresolved := *metadata
	unresolved.TransactionExtensions = []TransactionExtensionMetadata{{Identifier: "Missing", Type: 99}}
	_, err = unresolved.ExtensionsForVersion(0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unresolved type 99")
	_, err = (*ExtrinsicMetadata)(nil).ExtensionsForVersion(0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing extrinsic metadata")
	_, err = (&ExtrinsicMetadata{Versions: "05"}).ExtensionsForVersion(0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported transaction extension version")
}

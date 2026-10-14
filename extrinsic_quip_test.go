package scalecodec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/types/scaleBytes"
	"github.com/itering/scale.go/utiles"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

type quipSignatureManifest struct {
	Files map[string]struct {
		Bytes  int
		SHA256 string
		Method string
		Params []string
	}
	Blocks []struct {
		Number, ExecutionSpecVersion, PostStateSpecVersion, MetadataVersion, TransactionVersion int
		Hash, ParentHash, MetadataStateHash, MetadataFile, BlockFile                            string
		Extrinsics                                                                              []struct {
			Index, Version                                                              int
			Signed                                                                      bool
			Pallet, Call, CallIndexHex, CallSHA256, Hash, AccountID, Nonce, Tip, EraHex string
			EnvelopeSHA256, PublicKeySHA256, SignatureFieldSHA256                       string
		}
	}
}

func signatureFixtureRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func signatureFixtureSHA(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// These are the same byte fixtures used by the validator, Subxt and Apps
// conformance checks. No chain-specific registry overrides are installed.
func TestQuipCompositeSignatureFixtures(t *testing.T) {
	t.Cleanup(types.Clean)
	blocks, extrinsics, signed := 0, 0, 0
	for _, fixture := range []struct{ dir, sha string }{
		{"aglais-117-118", "a77d59562ecab6aaf80173e70a1f6dcf210815f8af471c7ce356e7ef08bd4b89"},
		{"local-119", "eab61fd6227b41dadc03de4df98ebc2c4d176506cb4c69124fa76b4ea7be4e48"},
	} {
		dir := filepath.Join("testdata", "quip-metadata-conformance", fixture.dir)
		manifestBytes := signatureFixtureRead(t, filepath.Join(dir, "manifest.json"))
		require.Equal(t, fixture.sha, signatureFixtureSHA(manifestBytes))
		var manifest quipSignatureManifest
		require.NoError(t, json.Unmarshal(manifestBytes, &manifest))
		for name, record := range manifest.Files {
			data := signatureFixtureRead(t, filepath.Join(dir, name))
			require.Equal(t, record.Bytes, len(data), name)
			require.Equal(t, record.SHA256, signatureFixtureSHA(data), name)
		}
		for _, block := range manifest.Blocks {
			blocks++
			t.Run(fmt.Sprintf("%s/%d", fixture.dir, block.Number), func(t *testing.T) {
				types.Clean()
				require.Equal(t, block.ParentHash, block.MetadataStateHash)
				require.Equal(t, "state_getMetadata", manifest.Files[block.MetadataFile].Method)
				require.Equal(t, []string{block.ParentHash}, manifest.Files[block.MetadataFile].Params)
				require.Equal(t, 7, block.TransactionVersion)
				metadataBytes := signatureFixtureRead(t, filepath.Join(dir, block.MetadataFile))
				metadata := MetadataDecoder{}
				metadata.Init(metadataBytes)
				metadata.Spec = block.ExecutionSpecVersion
				require.NoError(t, metadata.Process())
				require.Equal(t, len(metadataBytes), metadata.Data.Offset)
				require.Equal(t, 16, metadata.Metadata.MetadataVersion)
				var rawBlock struct {
					Block struct{ Extrinsics []string }
				}
				require.NoError(t, json.Unmarshal(signatureFixtureRead(t, filepath.Join(dir, block.BlockFile)), &rawBlock))
				require.Len(t, rawBlock.Block.Extrinsics, len(block.Extrinsics))
				for index, expected := range block.Extrinsics {
					extrinsics++
					require.Equal(t, index, expected.Index)
					input := utiles.HexToBytes(rawBlock.Block.Extrinsics[index])
					require.Equal(t, expected.Hash, fmt.Sprintf("0x%x", blake2b.Sum256(input)))
					if expected.Signed {
						signed++
					}
					var firstRaw interface{}
					for repeat := 0; repeat < 5; repeat++ {
						decoder := ExtrinsicDecoder{}
						option := &types.ScaleDecoderOption{Metadata: &metadata.Metadata, Spec: block.ExecutionSpecVersion}
						decoder.Init(scaleBytes.ScaleBytes{Data: input}, option)
						decoder.Process()
						require.Equal(t, len(input), decoder.Data.Offset)
						result := decoder.Value.(*GenericExtrinsic)
						require.Equal(t, fmt.Sprintf("%02x", expected.Version), result.VersionInfo)
						require.Equal(t, expected.Signed, decoder.ContainsTransaction)
						if expected.Pallet != "" {
							require.Equal(t, strings.ToLower(expected.Pallet), strings.ToLower(result.CallModule))
							require.Equal(t, expected.Call, result.CallModuleFunction)
						}
						require.Equal(t, strings.TrimPrefix(expected.CallIndexHex, "0x"), result.CallCode)
						require.Equal(t, expected.CallSHA256, signatureFixtureSHA(utiles.HexToBytes(result.CallCode+result.ParamsRaw)))
						if !expected.Signed {
							require.Empty(t, result.Signature)
							require.Nil(t, result.SignatureRaw)
							continue
						}
						require.Equal(t, expected.Hash, result.ExtrinsicHash)
						require.Equal(t, strings.TrimPrefix(expected.AccountID, "0x"), strings.TrimPrefix(fmt.Sprint(result.AccountId), "0x"))
						require.Equal(t, expected.Nonce, strconv.Itoa(result.Nonce))
						require.Equal(t, expected.Tip, result.Tip.String())
						require.Equal(t, strings.TrimPrefix(expected.EraHex, "0x"), result.Era)
						envelope, err := hex.DecodeString(result.Signature)
						require.NoError(t, err)
						require.Len(t, envelope, 1660)
						require.Equal(t, expected.EnvelopeSHA256, signatureFixtureSHA(envelope))
						require.Equal(t, expected.PublicKeySHA256, signatureFixtureSHA(envelope[:929]))
						require.Equal(t, expected.SignatureFieldSHA256, signatureFixtureSHA(envelope[929:]))
						// Independent wire oracle: compact length, version, then
						// MultiAddress::Id tag + 32-byte account precede the envelope.
						prefixBytes := len(input) - result.ExtrinsicLength
						start := prefixBytes + 1 + 1 + 32
						require.Equal(t, input[start:start+1660], envelope)
						require.IsType(t, map[string]interface{}{}, result.SignatureRaw)
						require.Equal(t, envelope, utiles.HexToBytes(types.EncodeWithOpt("ExtrinsicSignature", result.SignatureRaw, option)))
						if repeat == 0 {
							firstRaw = result.SignatureRaw
							for _, retained := range []int{1, 928, 929, 1659} {
								// Keep an accurate outer length while truncating inside
								// the public key or signature field of the real envelope.
								body := input[prefixBytes : start+retained]
								prefix := utiles.HexToBytes(types.Encode("Compact<U32>", len(body)))
								truncated := append(prefix, body...)
								negative := ExtrinsicDecoder{}
								negative.Init(scaleBytes.ScaleBytes{Data: truncated}, option)
								require.Panics(t, negative.Process, "extrinsic %d envelope cut at %d", index, retained)
							}
						}
						require.Equal(t, firstRaw, result.SignatureRaw)
					}
				}
			})
		}
	}
	require.Equal(t, 5, blocks)
	require.Equal(t, 25, extrinsics)
	require.Equal(t, 20, signed)
}

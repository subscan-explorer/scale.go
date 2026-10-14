# Shared metadata conformance inputs

The two fixture directories are byte-for-byte copies of
`quip-validator/docs/fixtures/metadata-conformance` at commit
`52a9070944d4cd2319a5cddcd2da1d8eb0a63abc`. Their manifests preserve capture
provenance, exact RPC queries, execution and post-state versions, and SHA-256
and byte counts for every file. No new capture or metadata conversion is used.

| Directory | Execution runtime | Blocks |
| --- | --- | --- |
| `aglais-117-118` | Archived 117/118, transaction version 7 | 330109, 330110, 591863, 591865 |
| `local-119` | Ordinary non-benchmark 119, transaction version 7 | 107 |

Metadata is the raw V16 result at the **parent block state**. Block 330109
executes under 117 and installs 118. The manifests pin the actual runtime code
hashes; archived source revisions and build features are not inferred.

The local fixture was captured from a disposable development node built at
`e5de2ddcb347565d15b75df823a116b67f5b602d` with default features, using
`SKIP_PALLET_REVIVE_FIXTURES=1 cargo build --offline -p quip-network-node`.
Its execution Wasm SHA-256 is
`d1126490ef240f287b4b691a1117216495b157d846b667712b6afc105f5aeaa8`.
The source repository contains capture/verification tooling in
`scripts/metadata-conformance/fixtures.py` and full reproduction instructions
in the fixture README at the pinned commit.

From this repository root, run:

```sh
go test . -run 'TestQuipCompositeSignatureFixtures|TestCompositeSignatureSpan|TestSignatureOutputConventions' -count=1
go test ./...
```

The shared-fixture test verifies both pinned manifest hashes and all file hashes,
then decodes all five blocks with their execution metadata, without Quip type
overrides. It checks complete consumption, order, raw extrinsic hashes, call
bytes, signer, nonce/era/tip, five bare V5 extrinsics and twenty signed V4
extrinsics. Every signed envelope must match the independent manifest's exact
1,660 bytes (929 public-key bytes plus 731 signature-field bytes), survive five
decodes unchanged, and re-encode from the preserved `SignatureRaw` value.
These sizes appear only in the fixture oracle, not the generic decoder.

Small synthetic tests cover nested composites, truncation, leading bytes that
could be mistaken for enum tags, a one-field struct, direct string signatures,
and the existing Ed25519/Sr25519/Ecdsa output convention. They exercise the
existing signed V4 and signed V5 decoder paths; signed V5 is separate from the
V5 General authorization format. Existing General tests remain in the suite.

Composite `signature` output is the complete consumed SCALE span as hex without
`0x`. Direct strings and enum string payloads retain their legacy output.
`signature_raw` remains the original metadata-decoded structure. This is decoder
coverage, not cryptographic verification or full indexer/database acceptance.
Event and runtime files are retained intact for shared consumer provenance;
this signature regression does not assert event ingestion or outcomes.

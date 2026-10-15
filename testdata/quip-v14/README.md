# V14 runtime type regression inputs

These are two unmodified request/response pairs from rr's read-only capture on
2026-10-14 at `https://bootnode-2.aglais.quip.network:20049/rpc`. The calls request
`Metadata_metadata_at_version(14)` at the parent states of shared fixture blocks
330109 (execution runtime 117) and 330110 (118). They contain the original
`Option<OpaqueMetadata>` wrapper. Runtime type IDs are **273** and **274**,
respectively; both resolve to `quip_protocol_runtime::Runtime` in their own
portable registry. IDs are specific to each metadata version and runtime.

`provenance.json` is an excerpt of the original capture manifest, retaining its
SHA-256, genesis identity, shared fixture manifest SHA-256, exact RPC params,
and request/response hashes. The source was
`/private/tmp/quip-subscan-rr-archive-capture-20261014/responses`; both pairs were
verified against that manifest before import. The tests pin the excerpt hash
and verify each raw request/response hash. No live RPC is required to test.

```sh
go test . -run '^TestMetadataV14(RuntimeType|QuipRuntimeType|DecoderProcess)$' -count=1
```

The synthetic case covers runtime type ID zero, absent/truncated IDs, an unknown
registry reference, and trailing bytes. Captured inputs check exact consumption,
the final compact ID, registry resolution, truncation, unknown references, and
rejection of extra trailing bytes. The top-level V14 metadata decoder enforces
exact consumption; the nested V14 codec consumes its own fields normally.
Existing decoder panic behavior is retained for malformed fields; trailing
bytes produce an error through `MetadataDecoder.Process`.

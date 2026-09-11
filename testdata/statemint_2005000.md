# Statemint runtime 2005000 regression fixtures

Captured on 2026-09-11 from public, finalized chain data during investigation of
17 `chain_blocks_20.codec_error=true` blocks. No API credentials are included.

- `statemint_2005000_v16.scale` is the exact SCALE response from
  `state_call("Metadata_metadata_at_version", "0x10000000", parent_hash)` at
  `0xb26089302dbbaab962750dba5195c1b45dff0fce5cd9896ffd700e2570197e5a`
  (parent of block 20499531). It includes the `Option<OpaqueMetadata>` wrapper.
- `statemint_2005000_extrinsics.json` contains all 17 failing general extrinsics
  fetched with `chain_getBlock`, plus the bare V5 timestamp and two signed V4
  control transactions from block 20503010.
- RPC endpoint: `wss://statemint.api.onfinality.io/public-ws`.
- Expected accounts were checked against the corresponding indexed
  `TransactionFeePaid.who` events. Expected nested calls, nonces, signatures and
  eras were recovered using the runtime's V16 version-one extension layout.
  Block 20499531-2 calls `Utility.force_batch` with `Staking.payout_stakers` for
  era 2288. Every fixture has `ChargeAssetTxPayment.asset_id=None` and tip zero.

The regression test uses the full metadata decoder, verifies complete byte
consumption, checks the call and authorization fields, computes the hash from
original bytes, and requires identical re-encoding. It does not make network
requests. A separate synthetic test covers version-zero/version-one ordering,
nonzero tips, missing values, and unsupported layouts.

The metadata contains view functions, exposing the former 8-byte/32-byte ID
mismatch. Its extension version 1 adds `VerifyMultiSignature`, `AsPgas`,
`AsDotnsGateway`, and `RestrictOrigins` to the older pipeline. The complete
catalog must not be flattened into `SignedExtensions`: V4 needs version zero,
while a general extrinsic must select its own version. Unknown versions cannot
safely fall back to that legacy layout.

Wire format reference:
https://github.com/paritytech/frame-metadata/blob/main/frame-metadata/src/v16.rs

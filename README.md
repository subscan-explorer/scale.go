# Scale Codec in Golang

Scale Codec written by golang, reference from [polkascan/py-scale-codec](https://github.com/polkascan/py-scale-codec).

## Features

- [x] Decode
- [x] Encode
- [x] Metadata decode(V11,V12,V13,V14)
- [x] Extrinsic decode
- [x] EventRecord decode
- [x] Custom type reg
- [x] Reg special version type

### Versioned transaction extensions

V5 general extrinsics (`0x45`) select their extension layout from V16 metadata
using `transaction_extension_version`. Decoding exposes the account and signature
from `VerifyMultiSignature`, along with nonce, era, tip, and the transaction hash.
The extension values are retained for byte-for-byte re-encoding, including `None`
options. V4 signed extrinsics use the version-zero layout.

Supply V16 metadata from the matching runtime when decoding a nonzero extension
version. V14/V15 metadata only describes version zero; unsupported versions and
invalid extension indexes fail explicitly instead of attempting the wrong layout.
As with other invalid SCALE input, `ExtrinsicDecoder.Process` reports decode
failures by panicking; `GenericExtrinsic.Encode` returns an error for missing or
unsupported extension layouts. V16 view function IDs are decoded as 32 bytes.

## Installation

```bash
go get -u github.com/itering/scale.go
```

## Test

```bash
go test ./... --cover -v 
```

Docker
```bash
docker build -t scale_go .
docker run -it scale_go
```

## Resources

- [Subscan](https://www.subscan.io)
- [Subscan Open-source](https://github.com/itering/subscan-essentials)
- [Polkadot.js](http://polkadot.js.org/)

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/itering/scale.go

## License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).

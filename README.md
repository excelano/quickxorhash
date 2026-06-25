# quickxorhash

A pure-Go implementation of Microsoft's QuickXorHash, the content hash that SharePoint Online and OneDrive for Business report for every file. It implements the standard library `hash.Hash` interface and has no dependencies beyond the Go standard library.

Microsoft Graph returns the digest for a file in the `file.hashes.quickXorHash` field of a driveItem. Computing the same hash locally lets a client tell whether a file already matches the copy stored in a document library without downloading it — the problem this package was written to solve, for an rsync-style SharePoint mirror that could not rely on timestamps surviving an upload.

## Install

```
go get github.com/excelano/quickxorhash
```

## Use

The value returned by `New` is a `hash.Hash`, so it streams like `crypto/sha1` or `hash/crc32`. Graph encodes the digest with standard base64, so compare against that:

```go
import (
	"encoding/base64"
	"io"
	"os"

	"github.com/excelano/quickxorhash"
)

func localHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := quickxorhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
```

For a one-shot hash of a byte slice already in memory, `quickxorhash.Sum` returns the raw 20-byte digest:

```go
sum := quickxorhash.Sum(data) // [20]byte
b64 := base64.StdEncoding.EncodeToString(sum[:])
```

## About the algorithm

QuickXorHash is a 160-bit hash that XORs the input into a rotating bit accumulator and folds the total length into the result. It is defined by Microsoft and described in the OneDrive developer documentation. This is an independent implementation from that description, verified against digests captured directly from Microsoft Graph (see the test vectors).

## License

MIT. Built by David M. Anderson with AI assistance (Claude, Anthropic).

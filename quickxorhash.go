// Package quickxorhash implements Microsoft's QuickXorHash, the content hash
// that SharePoint Online and OneDrive for Business report in a driveItem's
// file.hashes.quickXorHash field over Microsoft Graph. It lets a client decide
// whether a local file already matches the copy stored in a document library
// without downloading it — handy for sync, mirroring, and de-duplication.
//
// QuickXorHash is a 160-bit XOR-based hash defined by Microsoft and described in
// the OneDrive developer documentation. Graph returns the digest
// base64-encoded, so pair this package with encoding/base64 to compare:
//
//	h := quickxorhash.New()
//	if _, err := io.Copy(h, f); err != nil {
//		return err
//	}
//	got := base64.StdEncoding.EncodeToString(h.Sum(nil))
//	// compare got against driveItem.file.hashes.quickXorHash
//
// The value returned by New implements the standard library hash.Hash
// interface, so it streams and drops in anywhere crypto/sha1 or hash/crc32
// would. This is an independent implementation from the published algorithm
// description and depends only on the standard library.
package quickxorhash

import (
	"encoding/binary"
	"hash"
)

const (
	// Size is the length in bytes of a QuickXorHash digest.
	Size = 20
	// BlockSize is the hash's block size in bytes. The algorithm imposes no
	// real block size; this is the conventional value reported to callers.
	BlockSize = 64

	widthInBits    = 160
	shift          = 11
	bitsInLastCell = 32
)

// digest holds the running hash state. The 160-bit accumulator is kept as three
// 64-bit cells; only the low 32 bits of the third cell are used (32 + 64 + 64 =
// 160). shiftSoFar tracks the rotating bit offset so the hash streams correctly
// across successive Write calls of any size.
type digest struct {
	data        [3]uint64
	lengthSoFar uint64
	shiftSoFar  int
}

// New returns a new hash.Hash computing the QuickXorHash digest.
func New() hash.Hash { return &digest{} }

func (d *digest) Size() int      { return Size }
func (d *digest) BlockSize() int { return BlockSize }
func (d *digest) Reset()         { *d = digest{} }

// Write adds more data to the running hash. It never returns an error.
func (d *digest) Write(p []byte) (int, error) {
	cbSize := len(p)
	vectorArrayIndex := d.shiftSoFar / 64
	vectorOffset := d.shiftSoFar % 64

	iterations := cbSize
	if iterations > widthInBits {
		iterations = widthInBits
	}

	for i := 0; i < iterations; i++ {
		isLastCell := vectorArrayIndex == len(d.data)-1
		bitsInVectorCell := 64
		if isLastCell {
			bitsInVectorCell = bitsInLastCell
		}

		// When the byte fits inside the current cell it is XORed straight in;
		// otherwise it straddles two cells and is split across the boundary.
		if vectorOffset <= bitsInVectorCell-8 {
			for j := i; j < cbSize; j += widthInBits {
				d.data[vectorArrayIndex] ^= uint64(p[j]) << uint(vectorOffset)
			}
		} else {
			index1 := vectorArrayIndex
			index2 := 0
			if !isLastCell {
				index2 = vectorArrayIndex + 1
			}
			low := uint(bitsInVectorCell - vectorOffset)

			var xoredByte byte
			for j := i; j < cbSize; j += widthInBits {
				xoredByte ^= p[j]
			}
			d.data[index1] ^= uint64(xoredByte) << uint(vectorOffset)
			d.data[index2] ^= uint64(xoredByte) >> low
		}

		vectorOffset += shift
		for vectorOffset >= bitsInVectorCell {
			if isLastCell {
				vectorArrayIndex = 0
			} else {
				vectorArrayIndex++
			}
			vectorOffset -= bitsInVectorCell
		}
	}

	// Advance the rotating offset by the block length. Reducing cbSize modulo
	// widthInBits first keeps the running value identical to processing the
	// whole input at once, so the result is independent of how it was chunked.
	d.shiftSoFar = (d.shiftSoFar + shift*(cbSize%widthInBits)) % widthInBits
	d.lengthSoFar += uint64(cbSize)

	return cbSize, nil
}

// Sum appends the current 20-byte digest to b and returns the result. It does
// not change the underlying hash state.
func (d *digest) Sum(b []byte) []byte {
	s := d.checkSum()
	return append(b, s[:]...)
}

func (d *digest) checkSum() [Size]byte {
	var rgb [Size]byte
	binary.LittleEndian.PutUint64(rgb[0:8], d.data[0])
	binary.LittleEndian.PutUint64(rgb[8:16], d.data[1])

	// Only the low 4 bytes of the last cell belong in the 20-byte digest.
	var last [8]byte
	binary.LittleEndian.PutUint64(last[:], d.data[2])
	copy(rgb[16:20], last[0:4])

	// XOR the total length, little-endian, into the trailing 8 bytes.
	var lengthBytes [8]byte
	binary.LittleEndian.PutUint64(lengthBytes[:], d.lengthSoFar)
	for i := 0; i < len(lengthBytes); i++ {
		rgb[widthInBits/8-len(lengthBytes)+i] ^= lengthBytes[i]
	}
	return rgb
}

// Sum returns the QuickXorHash digest of data in a single call.
func Sum(data []byte) [Size]byte {
	d := &digest{}
	d.Write(data)
	return d.checkSum()
}

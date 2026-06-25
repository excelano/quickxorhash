package quickxorhash

import (
	"encoding/base64"
	"fmt"
	"hash"
	"testing"
)

// Both vectors were captured directly from Microsoft Graph: the bytes below
// were uploaded to a SharePoint Online document library and the expected
// strings are the quickXorHash values the service reported back. The 1000-byte
// case matters because it is the only one that exercises the inner striding
// loop and the cell wraparound, which never trigger for inputs under 160 bytes.
var vectors = []struct {
	name string
	data []byte
	want string // base64, as Graph returns it
}{
	{
		name: "empty",
		data: nil,
		want: "AAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	},
	{
		name: "23-byte",
		data: []byte("hello from xsync repro\n"),
		want: "CSQji5vp5jAfdiZvfwMI8DCHPLg=",
	},
	{
		name: "1000-byte",
		data: patterned(1000),
		want: "eelcfP1hi6r3o5C8MM9VxsvKe5Y=",
	},
}

// patterned returns a deterministic byte slice of length n. The same generator
// produced the bytes that were uploaded to capture the 1000-byte vector.
func patterned(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i*37 + 11) % 256)
	}
	return b
}

func sum(data []byte) string {
	s := Sum(data)
	return base64.StdEncoding.EncodeToString(s[:])
}

func TestKnownAnswers(t *testing.T) {
	for _, v := range vectors {
		if got := sum(v.data); got != v.want {
			t.Errorf("%s: Sum = %s, want %s", v.name, got, v.want)
		}
	}
}

// TestStreaming feeds each vector through Write in fixed-size chunks. A correct
// implementation produces the same digest no matter where the writes fall,
// including chunk sizes that straddle the 160-byte stride boundary.
func TestStreaming(t *testing.T) {
	chunks := []int{1, 3, 7, 13, 64, 160, 161, 333, 999}
	for _, v := range vectors {
		for _, c := range chunks {
			h := New()
			for off := 0; off < len(v.data); off += c {
				end := off + c
				if end > len(v.data) {
					end = len(v.data)
				}
				if _, err := h.Write(v.data[off:end]); err != nil {
					t.Fatalf("%s chunk=%d: Write: %v", v.name, c, err)
				}
			}
			got := base64.StdEncoding.EncodeToString(h.Sum(nil))
			if got != v.want {
				t.Errorf("%s chunk=%d: %s, want %s", v.name, c, got, v.want)
			}
		}
	}
}

func TestSumDoesNotMutate(t *testing.T) {
	h := New()
	h.Write([]byte("hello from xsync repro\n"))
	first := base64.StdEncoding.EncodeToString(h.Sum(nil))
	second := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if first != second {
		t.Errorf("Sum mutated state: %s then %s", first, second)
	}
}

func TestReset(t *testing.T) {
	h := New()
	h.Write(patterned(500))
	h.Reset()
	h.Write([]byte("hello from xsync repro\n"))
	if got := base64.StdEncoding.EncodeToString(h.Sum(nil)); got != "CSQji5vp5jAfdiZvfwMI8DCHPLg=" {
		t.Errorf("after Reset: %s, want CSQji5vp5jAfdiZvfwMI8DCHPLg=", got)
	}
}

func TestSizes(t *testing.T) {
	h := New()
	if h.Size() != Size || h.Size() != 20 {
		t.Errorf("Size = %d, want 20", h.Size())
	}
	if h.BlockSize() != BlockSize {
		t.Errorf("BlockSize = %d, want %d", h.BlockSize(), BlockSize)
	}
}

// Compile-time confirmation that New returns a usable hash.Hash.
var _ hash.Hash = New()

func ExampleNew() {
	h := New()
	h.Write([]byte("hello from xsync repro\n"))
	fmt.Println(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	// Output: CSQji5vp5jAfdiZvfwMI8DCHPLg=
}

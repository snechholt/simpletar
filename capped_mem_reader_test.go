package simpletar

import (
	"bytes"
	"github.com/snechholt/simplefs"
	"testing"
)

func TestCappedMemReader(t *testing.T) {
	t.Run("Handle in memory", func(t *testing.T) {
		const limit = 999999
		var fs simplefs.MemFS
		testReader(t, func(t *testing.T, b []byte) Reader {
			r, err := CappedMemReader(bytes.NewReader(b), limit, &fs)
			if err != nil {
				t.Fatal(err)
			}
			return r
		})
		if fs.Size() != 0 {
			t.Fatalf("No files should have been written to disk")
		}
	})

	t.Run("Write to disk", func(t *testing.T) {
		const limit = 0
		var fs simplefs.MemFS
		testReader(t, func(t *testing.T, b []byte) Reader {
			r, err := CappedMemReader(bytes.NewReader(b), limit, &fs)
			if err != nil {
				t.Fatal(err)
			}
			return r
		})
		if fs.Size() == 0 {
			t.Fatalf("Files should have been written to disk")
		}
	})
}

package simpletar

import (
	"bytes"
	"github.com/snechholt/simplefs"
	"io/ioutil"
	"testing"
)

func TestUntarReader(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		testReader(t, func(t *testing.T, b []byte) Reader {
			r, err := UntarReader(bytes.NewReader(b), &simplefs.MemFS{}, nil)
			if err != nil {
				t.Fatal(err)
			}
			return r
		})
	})
	t.Run("options.gzip = false", func(t *testing.T) {
		testReader(t, func(t *testing.T, b []byte) Reader {
			options := &UntarOptions{Gzip: false}
			r, err := UntarReader(bytes.NewReader(b), &simplefs.MemFS{}, options)
			if err != nil {
				t.Fatal(err)
			}
			return r
		})
	})
	t.Run("options.gzip = true", func(t *testing.T) {
		testReader(t, func(t *testing.T, b []byte) Reader {
			options := &UntarOptions{Gzip: true}
			r, err := UntarReader(bytes.NewReader(b), &simplefs.MemFS{}, options)
			if err != nil {
				t.Fatal(err)
			}
			return r
		})
	})
}

func testReader(t *testing.T, ctor func(t *testing.T, b []byte) Reader) {
	files := []testFile{
		{Name: "A", B: []byte{1, 2, 3}},
		{Name: "B", B: []byte{2, 3, 4, 5}},
		{Name: "C", B: []byte{3, 4, 5, 6, 7}},
	}
	tests := map[string]bool{
		"gzipped":     true,
		"not gzipped": false,
	}
	for name, gz := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			w := &Writer{W: &buf, Gzip: gz}
			for _, file := range files {
				if _, err := w.Write(file.Name, file.B); err != nil {
					t.Fatal(err)
				}
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			tr := ctor(t, buf.Bytes())
			for _, file := range files {
				r, err := tr.Open(file.Name)
				if err != nil {
					t.Fatalf("Open(%s) error: %v", file.Name, err)
				}
				b, err := ioutil.ReadAll(r)
				if err != nil {
					t.Fatalf("Open(%s).ReadAll() error : %v", file.Name, err)
				}
				if bytes.Compare(file.B, b) != 0 {
					t.Errorf("Wrong file contents in file %s: want %v, got %v", file.Name, file.B, b)
				}
				if err := r.Close(); err != nil {
					t.Fatalf("Open(%s).Close() error: %v", file.Name, err)
				}
			}
		})
	}
}

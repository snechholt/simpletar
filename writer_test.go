package simpletar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/snechholt/simplefs"
	"io"
	"io/ioutil"
	"testing"
)

func TestWriter(t *testing.T) {
	files := []testFile{
		{Name: "A", B: []byte{1, 2, 3}},
		{Name: "B", B: []byte{2, 3, 4, 5}},
		{Name: "C", B: []byte{3, 4, 5, 6, 7}},
	}

	tests := []struct {
		Name   string
		W      *Writer
		Method string
	}{
		{Name: "Write()", W: &Writer{}, Method: "Write"},
		{Name: "Write() [gzip]", W: &Writer{Gzip: true}, Method: "Write"},

		{Name: "Create()", W: &Writer{}, Method: "Create"},
		{Name: "Create() [gzip]", W: &Writer{Gzip: true}, Method: "Create"},
		{Name: "Create() [fs]", W: &Writer{FS: &simplefs.MemFS{}}, Method: "Create"},
		{Name: "Create() [gzip,fs]", W: &Writer{Gzip: true, FS: &simplefs.MemFS{}}, Method: "Create"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var buf bytes.Buffer
			test.W.W = &buf
			for _, file := range files {
				switch test.Method {
				case "Write":
					if _, err := test.W.Write(file.Name, file.B); err != nil {
						t.Fatalf("Write(%s) error: %v", file.Name, err)
					}
				case "Create":
					w := test.W.Create(file.Name)
					if _, err := w.Write(file.B); err != nil {
						t.Fatalf("Create(%s).Write() error: %v", file.Name, err)
					}
					if err := w.Close(); err != nil {
						t.Fatalf("Create(%s).Close() error: %v", file.Name, err)
					}
				}
			}
			if err := test.W.Close(); err != nil {
				t.Fatalf("Close() error: %v", err)
			}
			validateTarball(t, &buf, test.W.Gzip, files)
		})
	}

}

type testFile struct {
	Name string
	B    []byte
}

type testFileSlice []testFile

func (s testFileSlice) Find(name string) (testFile, bool) {
	for _, f := range s {
		if f.Name == name {
			return f, true
		}
	}
	return testFile{}, false
}

func validateTarball(t *testing.T, r io.Reader, gz bool, want testFileSlice) {
	if gz {
		var err error
		if r, err = gzip.NewReader(r); err != nil {
			t.Fatal(err)
		}
	}
	tr := tar.NewReader(r)

	var got testFileSlice
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		f := testFile{Name: hdr.Name, B: b}
		got = append(got, f)
	}
	for _, wantFile := range want {
		gotFile, ok := got.Find(wantFile.Name)
		if !ok {
			t.Errorf("Missing file: %v", wantFile.Name)
		} else if bytes.Compare(wantFile.B, gotFile.B) != 0 {
			t.Errorf("Wrong content for %s: want %v, got %v", wantFile.Name, wantFile.B, gotFile.B)
		}
	}
	for _, file := range got {
		if _, ok := want.Find(file.Name); !ok {
			t.Errorf("Unexpected file: %s", file.Name)
		}
	}
}

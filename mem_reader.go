package simpletar

import (
	"archive/tar"
	"bytes"
	"github.com/snechholt/simplefs"
	"io"
	"io/ioutil"
)

// MemReader returns a Reader that reads from r and stores the tarball
// contents in-memory. If parameter eager is true, each of the tarball's
// files will be untar'ed and stored as an internal map, otherwise the
// tarball itself will be stored in-memory and untar'ed on every call to
// Open.
func MemReader(r io.Reader, eager ...bool) (Reader, error) {
	if lazy := len(eager) == 0 || !eager[0]; lazy {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return &lazyMemReader{b: b}, nil
	}
	src, err := open(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()
	tr := tar.NewReader(src)
	m := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		m[hdr.Name] = b
	}
	return &eagerMemReader{m: m}, nil
}

type lazyMemReader struct {
	b []byte
}

func (reader *lazyMemReader) Open(name string) (simplefs.File, error) {
	openFn := func() (io.ReadCloser, error) {
		return &readCloser{r: bytes.NewReader(reader.b)}, nil
	}
	sr := StreamReader(openFn)
	return sr.Open(name)
}

type eagerMemReader struct {
	m map[string][]byte
}

func (reader *eagerMemReader) Open(name string) (simplefs.File, error) {
	b, ok := reader.m[name]
	if !ok {
		return nil, simplefs.ErrNotFound
	}
	return &file{r: bytes.NewReader(b)}, nil
}

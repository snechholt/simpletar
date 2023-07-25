package simpletar

import (
	"archive/tar"
	"github.com/snechholt/simplefs"
	"io"
)

type streamReader struct {
	openFn func() (io.ReadCloser, error)
}

// StreamReader returns a Reader that open and reads from the provided function
// every time Open is called.
func StreamReader(openFn func() (io.ReadCloser, error)) Reader {
	return &streamReader{openFn: openFn}
}

func (reader *streamReader) Open(name string) (simplefs.File, error) {
	var found bool

	r, err := reader.openFn()
	if err != nil {
		return nil, err
	}
	defer func() {
		if !found {
			_ = r.Close()
		}
	}()

	src, err := open(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !found {
			_ = src.Close()
		}
	}()

	tr := tar.NewReader(src)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == name {
			found = true
			closeFn := func() error {
				if err := src.Close(); err != nil {
					return err
				}
				if err := r.Close(); err != nil {
					return err
				}
				return nil
			}
			return &file{r: tr, closeFn: closeFn}, nil
		}
	}

	return nil, simplefs.ErrNotFound
}

func (reader *streamReader) ForEachFile(fn func(name string, r io.Reader) error) error {
	r, err := reader.openFn()
	if err != nil {
		return err
	}
	defer r.Close()

	src, err := open(r)
	if err != nil {
		return err
	}
	defer src.Close()

	tr := tar.NewReader(src)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := fn(hdr.Name, tr); err != nil {
			return err
		}
	}

	return nil
}

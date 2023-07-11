package simpletar

import (
	"bytes"
	"github.com/snechholt/simplefs"
	"io"
)

// CappedMemReader returns a Reader where the implementation depends on the size of the tarball. If
// the size is less than capacity, MemReader is used, otherwise UntarReader is used.
func CappedMemReader(r io.Reader, capacity int64, fs simplefs.FS, untarOptions ...*UntarOptions) (Reader, error) {
	var buf bytes.Buffer
	_, err := io.CopyN(&buf, r, capacity)
	if err == io.EOF {
		return &lazyMemReader{b: buf.Bytes()}, nil
	}
	if err != nil {
		return nil, err
	}
	mr := io.MultiReader(&buf, r)
	var opt *UntarOptions
	if len(untarOptions) > 0 {
		opt = untarOptions[0]
	}
	return UntarReader(mr, fs, opt)
}

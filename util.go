package simpletar

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/snechholt/simplefs"
	"io"
)

func open(r io.Reader) (io.ReadCloser, error) {
	const (
		gzipID1 = 0x1f
		gzipID2 = 0x8b
	)
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	mw := io.MultiReader(bytes.NewBuffer(buf), r)
	if isGzip := buf[0] == gzipID1 && buf[1] == gzipID2; isGzip {
		gr, err := gzip.NewReader(mw)
		if err != nil {
			return nil, err
		}
		return gr, nil
	}
	return &readCloser{r: mw}, nil
}

type file struct {
	r       io.Reader
	closeFn func() error
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *file) Close() error {
	if f.closeFn != nil {
		return f.closeFn()
	}
	return nil
}

func (f *file) ReadDir(n int) ([]simplefs.DirEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

type gzippedFile struct {
	f  simplefs.File
	gr *gzip.Reader
}

func (f *gzippedFile) Read(p []byte) (n int, err error) {
	if f.gr == nil {
		gr, err := gzip.NewReader(f.f)
		if err != nil {
			return 0, err
		}
		f.gr = gr
	}
	return f.gr.Read(p)
}

func (f *gzippedFile) Close() error {
	var err error
	if f.gr != nil {
		err = f.gr.Close()
	}
	if f.f != nil {
		err = f.f.Close()
	}
	return err
}

func (f *gzippedFile) ReadDir(n int) ([]simplefs.DirEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

type readCloser struct {
	r       io.Reader
	closeFn func() error
}

// Read reads from the underlying reader into p.
func (r *readCloser) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

// Close calls the close function provided when creating the reader.
func (r *readCloser) Close() error {
	if r.closeFn != nil {
		return r.closeFn()
	}
	return nil
}

type writeCloser struct {
	w       io.Writer
	closeFn func() error
}

func (w *writeCloser) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writeCloser) Close() error {
	if w.closeFn != nil {
		return w.closeFn()
	}
	return nil
}

type errReader struct {
	n   int
	err error
}

func (w *errReader) Read(p []byte) (int, error) {
	return w.n, w.err
}

type errWriter struct {
	n   int
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return w.n, w.err
}

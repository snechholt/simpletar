package simpletar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/snechholt/simplefs"
	"io"
)

type Writer struct {
	W    io.Writer
	FS   simplefs.FS
	Gzip bool

	tw     *tar.Writer
	gw     *gzip.Writer
	buffer []tar.Header
}

func (w *Writer) init() {
	if w.gw == nil && w.Gzip {
		w.gw = gzip.NewWriter(w.W)
	}
	if w.tw == nil {
		if w.Gzip {
			w.tw = tar.NewWriter(w.gw)
		} else {
			w.tw = tar.NewWriter(w.W)
		}
	}
}

// TODO: add optional param n int64, specifying the length so that we can write directly to tr
func (w *Writer) Create(name string) io.WriteCloser {
	if w.FS == nil {
		var buf bytes.Buffer
		return &writeCloser{
			w: &buf,
			closeFn: func() error {
				_, err := w.Write(name, buf.Bytes())
				return err
			},
		}
	}

	f, err := w.FS.Create(name)
	if err != nil {
		return &writeCloser{w: &errWriter{err: err}}
	}

	dst := f
	var gw *gzip.Writer
	if w.Gzip {
		gw = gzip.NewWriter(f)
		dst = gw
	}

	var sw sizeWriter
	mw := io.MultiWriter(dst, &sw)

	return &writeCloser{
		w: mw,
		closeFn: func() error {
			if gw != nil {
				if err := gw.Close(); err != nil {
					return err
				}
			}
			if err := f.Close(); err != nil {
				return err
			}
			bf := tar.Header{Name: name, Size: sw.size}
			w.buffer = append(w.buffer, bf)
			return nil
		},
	}
}

func (w *Writer) Write(name string, b []byte) (n int, err error) {
	w.init()
	hdr := &tar.Header{Name: name, Size: int64(len(b))}
	if err := w.tw.WriteHeader(hdr); err != nil {
		return 0, err
	}
	return w.tw.Write(b)
}

func (w *Writer) Close() error {
	w.init()
	for _, hdr := range w.buffer {
		err := func() error {
			if err := w.tw.WriteHeader(&hdr); err != nil {
				return err
			}
			r, err := w.FS.Open(hdr.Name)
			if err != nil {
				return err
			}
			defer func() { _ = r.Close() }()
			var src io.Reader = r
			if w.Gzip {
				gr, err := gzip.NewReader(r)
				if err != nil {
					return err
				}
				defer func() { _ = gr.Close() }()
				src = gr
			}
			if _, err := io.Copy(w.tw, src); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	if err := w.tw.Close(); err != nil {
		return err
	}
	if w.gw != nil {
		if err := w.gw.Close(); err != nil {
			return err
		}
	}
	return nil
}

type sizeWriter struct {
	size int64
}

func (w *sizeWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.size += int64(n)
	return n, nil
}

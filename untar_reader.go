package simpletar

import (
	"archive/tar"
	"compress/gzip"
	"github.com/snechholt/simplefs"
	"io"
)

type untarReader struct {
	fs      simplefs.FS
	gzipped bool
}

// UntarOptions specifies how UntarReader handles a tarball.
type UntarOptions struct {
	// Gzip specifies whether to gzip the file contents on the destination
	// file system when the tarball's files are untar'ed.
	Gzip bool
}

func (options *UntarOptions) getGzip() bool {
	return options != nil && options.Gzip
}

// UntarReader returns a Reader that untars the tarballs contents to fs and then
// accesses the files through fs when they are opened.
func UntarReader(r io.Reader, fs simplefs.FS, options *UntarOptions) (Reader, error) {
	doGzip := options.getGzip()
	src, err := open(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()
	tr := tar.NewReader(src)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := hdr.Name
		if doGzip {
			name += ".gz"
		}
		var w io.WriteCloser
		w, err = fs.Create(name)
		if err != nil {
			return nil, err
		}
		if options.getGzip() {
			gw := gzip.NewWriter(w)
			_w := w
			w = &writeCloser{
				w: gw,
				closeFn: func() error {
					if err := gw.Close(); err != nil {
						return err
					}
					return _w.Close()
				},
			}
		}
		if _, err := io.Copy(w, tr); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	}
	return &untarReader{fs: fs, gzipped: doGzip}, nil
}

func (reader *untarReader) Open(name string) (simplefs.File, error) {
	if reader.gzipped {
		name += ".gz"
	}
	r, err := reader.fs.Open(name)
	if err != nil {
		return nil, err
	}
	if reader.gzipped {
		return &gzippedFile{f: r}, nil
	}
	return r, nil
}

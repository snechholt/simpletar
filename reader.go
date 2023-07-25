package simpletar

import (
	"github.com/snechholt/simplefs"
	"io"
)

// Reader provides an interface for reading a tarball and accessing its contents.
type Reader interface {
	// Open searches for the specified file name within the tarball. If no file with
	// the specified name was found, it returns ErrNotFound.
	Open(name string) (simplefs.File, error)

	// ForEachFile iterates over each file in the tarball, calling fn for each file.
	// If fn returns an error, the iteration stops and the error is returned to the
	// caller.
	ForEachFile(fn func(name string, r io.Reader) error) error
}

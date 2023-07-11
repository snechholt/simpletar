package simpletar

import (
	"github.com/snechholt/simplefs"
)

// Reader provides an interface for reading a tarball and accessing its contents.
type Reader interface {
	// Open searches for the specified file name within the tarball. If no file with
	// the specified name was found, it returns ErrNotFound.
	Open(name string) (simplefs.File, error)
}

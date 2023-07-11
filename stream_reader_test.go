package simpletar

import (
	"bytes"
	"io"
	"testing"
)

func TestStreamReader(t *testing.T) {
	testReader(t, func(t *testing.T, b []byte) Reader {
		openFn := func() (io.ReadCloser, error) {
			return &readCloser{r: bytes.NewReader(b)}, nil
		}
		return StreamReader(openFn)
	})
}

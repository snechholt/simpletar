package simpletar

import (
	"bytes"
	"testing"
)

func TestLazyMemReader(t *testing.T) {
	testReader(t, func(t *testing.T, b []byte) Reader {
		r, err := MemReader(bytes.NewReader(b), true)
		if err != nil {
			t.Fatal(err)
		}
		return r
	})
}

func TestEagerMemReader(t *testing.T) {
	testReader(t, func(t *testing.T, b []byte) Reader {
		r, err := MemReader(bytes.NewReader(b), false)
		if err != nil {
			t.Fatal(err)
		}
		return r
	})
}

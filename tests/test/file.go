package test

import (
	"io/ioutil"
	"testing"
)

func ReadFile(t *testing.T, path string) []byte {
	t.Helper()

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error file reading file %q: %#v", path, err)
	}

	return bs
}

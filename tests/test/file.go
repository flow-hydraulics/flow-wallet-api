package test

import (
	"os"
	"testing"
)

func ReadFile(t *testing.T, path string) []byte {
	t.Helper()

	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("error file reading file %q: %#v", path, err)
	}

	return bs
}

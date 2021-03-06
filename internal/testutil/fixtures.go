package testutil

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func MustSpecFixture(absPath string) []byte {
	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		panic(fmt.Sprintf("error loading fixture %s: %v", absPath, err))
	}

	return bytes
}

func SpecFixture(t *testing.T, relPath string) []byte {
	t.Helper()

	return Fixture(t, filepath.Join("message-api-spec", relPath))
}

func Fixture(t *testing.T, relPath string) []byte {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("error loading caller")
	}

	p := path.Join(path.Dir(filename), "../../", relPath)

	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		t.Fatalf("error loading fixture %s: %v", p, err)
	}

	return bytes
}

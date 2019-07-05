package app

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestMainHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"rdss-archivematica-channel-adapter", "help"}

	var (
		output    bytes.Buffer
		errOutput bytes.Buffer
	)
	err := Run(&output, &errOutput)

	if err != nil {
		t.Error(err)
	}
	if have, want := output.String(), "Available Commands"; !strings.Contains(have, want) {
		t.Errorf("expected output %s not found in output: %s", want, have)
	}
	if errOutput.String() != "" {
		t.Errorf("error output is not empty")
	}
}

func TestMainUnknownCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"rdss-archivematica-channel-adapter", "unknown"}

	err := Run(ioutil.Discard, ioutil.Discard)

	if err == nil {
		t.Error("error expected")
	}
}

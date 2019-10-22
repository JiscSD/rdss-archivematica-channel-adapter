package message_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
)

// TestVersion expects the nested repository (message-api-spec) to be pre-downloaded.
// Use `git submodule update` to download the version set in the root repo.
func TestVersion(t *testing.T) {
	t.Skip("Not using tag v4.0.0 - https://github.com/JiscSD/rdss-message-api-specification/pull/124 not released yet")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	cmd.Dir = filepath.Join(wd, "../../message-api-spec")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if have, want := strings.TrimSpace(string(out)), message.Version; have != want {
		t.Fatalf("wanted %v; got %v", []byte(have), []byte(want))
	}
}

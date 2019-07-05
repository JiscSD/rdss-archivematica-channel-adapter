// +build integration

package adapter

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

type Runner struct {
	command    string
	configFile string
	dir        string
	args       []string
	env        []string
}

func Server(args ...string) *Runner {
	return &Runner{command: "server", args: args}
}

func (b *Runner) RunOrFail(t *testing.T) {
	t.Helper()
	if err := b.Run(t); err != nil {
		t.Fatal(err)
	}
}

func (b *Runner) Run(t *testing.T) error {
	t.Helper()

	cmd := b.exec(context.Background())
	t.Log(cmd.Args)

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "rdss-archivematica-channel-adapter %s", b.command)
	}

	t.Logf("Start in %s", time.Since(start))
	return nil
}

func (b *Runner) exec(ctx context.Context) *exec.Cmd {
	args := []string{b.command}
	if b.configFile != "" {
		args = append(args, "--config", b.configFile)
	}
	args = append(args, b.args...)

	cmd := exec.CommandContext(ctx, "rdss-archivematica-channel-adapter", args...)
	cmd.Env = append(removeAppEnvs(os.Environ()), b.env...)
	if b.dir != "" {
		cmd.Dir = b.dir
	}

	// If the test is killed by a timeout, go test will wait for
	// os.Stderr and os.Stdout to close as a result.
	//
	// However, the `cmd` will still run in the background
	// and hold those descriptors open.
	// As a result, go test will hang forever.
	//
	// Avoid that by wrapping stderr and stdout, breaking the short
	// circuit and forcing cmd.Run to use another pipe and goroutine
	// to pass along stderr and stdout.
	// See https://github.com/golang/go/issues/23019
	cmd.Stdout = struct{ io.Writer }{os.Stdout}
	cmd.Stderr = struct{ io.Writer }{os.Stderr}

	return cmd
}

func removeAppEnvs(env []string) []string {
	var clean []string

	for _, value := range env {
		if !strings.HasPrefix(value, "RDSS_ARCHIVEMATICA_ADAPTER_") {
			clean = append(clean, value)
		}
	}

	return clean
}

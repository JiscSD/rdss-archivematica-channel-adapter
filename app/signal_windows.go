// +build windows

package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/adapter"

	"github.com/pkg/errors"
)

func interrupt(cancel <-chan struct{}, registry *adapter.Registry) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-cancel:
		return errors.New("canceled")
	}
}

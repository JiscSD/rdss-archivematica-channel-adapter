package main

import (
	"context"
	"os"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/app"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := app.Run(os.Stdout, os.Stderr); err != nil {
		if errors.Cause(err) == context.Canceled {
			logrus.Debugln(errors.Wrap(err, "ignore error since context is cancelled"))
		} else {
			logrus.Fatal(err)
		}
	}
}

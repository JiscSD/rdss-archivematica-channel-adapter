// +build tools

package tools

import (
	// Install with `make tools`
	_ "github.com/johnewart/io-bindata"
	_ "golang.org/x/tools/cmd/cover"

	// Go 1.12 issue
	_ "github.com/go-logfmt/logfmt"
)

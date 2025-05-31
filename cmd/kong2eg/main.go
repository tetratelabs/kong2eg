// Copyright (c) Tetrate, Inc All Rights Reserved.

package main

import (
	"fmt"
	"os"

	"github.com/tetratelabs/kong2eg/internal/cmd/kong2eg"
)

func main() {
	if err := kong2eg.GetRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

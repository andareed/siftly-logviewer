package main

import (
	"fmt"
	"os"

	"github.com/andareed/siftly-hostlog/internal/pluginlog/cli"
)

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

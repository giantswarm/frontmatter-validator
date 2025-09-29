package main

import (
	"os"

	"github.com/giantswarm/frontmatter-validator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"os"

	"github.com/sirsjg/momentum/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

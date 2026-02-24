package main

import (
	"os"

	"github.com/stephenmfriend/momentum/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

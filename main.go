package main

import (
	"os"

	"github.com/stevegrehan/momentum/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

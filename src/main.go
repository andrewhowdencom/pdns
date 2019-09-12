package main

import (
	"os"

	"github.com/dedelala/sysexits"
)

func main() {
	os.Exit(sysexits.Unavailable)
}
package main

import (
	"os"

	"github.com/adamdost-0/hash-engine/internal/hashctl"
)

func main() {
	os.Exit(hashctl.Run(os.Args[1:], os.Stdout, os.Stderr))
}

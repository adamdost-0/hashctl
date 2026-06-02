package main

import (
	"os"

	"github.com/adamdost-0/hashctl/internal/hashctl"
)

func main() {
	os.Exit(hashctl.Run(os.Args[1:], os.Stdout, os.Stderr))
}

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: migrate up")
		os.Exit(2)
	}

	// The migration driver and first approved schema are intentionally pending.
	// Keep the command explicit so deployment fails loudly once migrations are enabled.
	fmt.Fprintln(os.Stderr, "migration runner is not configured yet")
	os.Exit(1)
}

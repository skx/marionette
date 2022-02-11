//go:build !go1.18
// +build !go1.18

package main

import (
	"fmt"
)

var (
	version = "unreleased"
)

func showVersion() {
	fmt.Printf("%s\n", version)
}

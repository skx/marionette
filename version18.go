//go:build go1.18
// +build go1.18

package main

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	version = "unreleased"
)

func showVersion() {
	fmt.Printf("%s\n", version)

	info, ok := debug.ReadBuildInfo()

	if ok {
		for _, settings := range info.Settings {
			if strings.Contains(settings.Key, "vcs") {
				fmt.Printf("%s: %s\n", settings.Key, settings.Value)
			}
		}
	}

}

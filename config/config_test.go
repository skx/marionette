// Package config holds global options.
//
// Options are intended to be set via the command-line flags,
// and made available to all our plugins.
package config

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {

	// Create an object
	c := &Config{Debug: false, Verbose: false}

	// Verify the options are sane
	if c.Debug != false {
		t.Fatalf("structure test failed")
	}
	if c.Verbose != false {
		t.Fatalf("structure test failed")
	}

	// Convert to string
	out := c.String()
	if !strings.Contains(out, "Debug:false") {
		t.Fatalf("string output has wrong content")
	}
	if !strings.Contains(out, "Verbose:false") {
		t.Fatalf("string output has wrong content")
	}

	// Change settings
	c.Debug = true
	out = c.String()
	if !strings.Contains(out, "Debug:true") {
		t.Fatalf("string output has wrong content")
	}
	if !strings.Contains(out, "Verbose:false") {
		t.Fatalf("string output has wrong content")
	}

	// Change settings
	c.Verbose = true
	out = c.String()
	if !strings.Contains(out, "Debug:true") {
		t.Fatalf("string output has wrong content")
	}
	if !strings.Contains(out, "Verbose:true") {
		t.Fatalf("string output has wrong content")
	}

	// Nil-test
	c = nil
	out = c.String()
	if out != "Config{nil}" {
		t.Fatalf("string output has wrong content for nil object")
	}
}

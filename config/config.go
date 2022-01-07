// Package config holds global options.
//
// Options are intended to be set via the command-line flags,
// and made available to all our plugins.
package config

// Config holds the state which is set by the main driver, and is
// made available to all of our plugins.
type Config struct {

	// Debug is used to let our plugins know that the marionette
	// CLI was started with the `-debug` flag present.
	Debug bool

	// Verbose is used to let our plugins know that the marionette
	// CLI was started with the `-verbose` flag present.
	Verbose bool
}

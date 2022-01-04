// Package config holds global options.
//
// Options are intended to be set via the command-line flags,
// and made available to all our plugins.
package config

// Config holds the state which is set by the main driver, and is
// made available to all of our plugins.
type Config struct {

	// Verbose is the only configuration option at the moment,
	// and controls whether we should be quiet, or noisy, when
	// processing our rule-files.
	Verbose bool
}

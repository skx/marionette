# Examples

This directory contains some simple examples, to show how this tool can be used.

* [install-go.recipe](install-go.recipe)
  * This shows downloading a binary distribution of the golang compiler/toolset.
  * The binary release is downloaded beneath `/opt/go-${version}`.
  * A symlink is created to make version upgrades simple.

* [433-mhz-temperature-grapher-docker.recipe](433-mhz-temperature-grapher-docker.recipe)
  * This example writes a `docker-compose.yml` file into a directory.
  * It also pulls three docker containers from their remote source.
  * Finally it restarts the application.

* [download.recipe](download.recipe)
  * Demonstrates how to use the conditional assignments to choose programs.
  * Selecting either wget or curl, depending on which is available.

* [overview.recipe](overview.recipe)
  * This is a well-commented example that shows numerous examples of the various modules.

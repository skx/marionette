#!/bin/sh

# Failures will cause this script to terminate
set -e

# I don't even ..
go env -w GOFLAGS="-buildvcs=false"

# Run the tests
go test -race ./...

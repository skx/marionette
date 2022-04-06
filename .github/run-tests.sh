#!/bin/sh

# Failures will cause this script to terminate
set -e

# Run the tests
go test -race ./...

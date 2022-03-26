#!/bin/sh

# Run golang tests
echo "Running our project-specific test-cases .."
go test -race ./...
echo "Completed our project-specific test-cases .."

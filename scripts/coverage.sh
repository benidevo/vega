#!/bin/sh

# Run tests and generate coverage profile
go test ./... -coverprofile=coverage.out -coverpkg='github.com/benidevo/vega/internal/...' -covermode=set

# Filter out handlers from the coverage output
grep -v '/handlers.go' coverage.out > coverage.filtered.out

# Generate report
if [ "$1" = "verbose" ]; then
  go tool cover -func=coverage.filtered.out
else
  go tool cover -func=coverage.filtered.out | grep total:
fi
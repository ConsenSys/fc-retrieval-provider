#!/bin/bash

trgt_coverage=${1:-"80"}

go test ./... -coverprofile cover.out -coverpkg=./...
curr_coverage=$(go tool cover -func cover.out | grep total | awk '{print $3}' | tr -d '%')

echo "Total: $curr_coverage%"

if awk "BEGIN {exit !("$curr_coverage" >= "$trgt_coverage")}"; then
  echo "Unit tests are passing $trgt_coverage% coverage"
  exit 0
else
  echo "Unit tests do not pass $trgt_coverage% coverage"
  exit 1
fi
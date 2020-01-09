#!/usr/bin/env bash
set -e

# aborts build (via exit 1) if srcclr scan detects vulnerabilities

RESULTS=$(srcclr scan . --json | jq '.records[].vulnerabilities')
NUM_RESULTS=$(echo "$RESULTS" | jq '.|length')

if [[ "$NUM_RESULTS" != "0" ]]; then
  echo "SRCCLR detected $NUM_RESULTS vulnerabilities"
  echo "$RESULTS"
  exit 1
fi

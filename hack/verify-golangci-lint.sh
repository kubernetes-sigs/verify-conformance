#!/bin/bash

set -o errexit
set -o nounset

cd "$(git rev-parse --show-toplevel)" || exit 1

golangci-lint run

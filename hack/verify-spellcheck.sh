#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

go install github.com/client9/misspell/cmd/misspell@latest
misspell -error main.go cmd/ pkg/ docs/ config/ README*

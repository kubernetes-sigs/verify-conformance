#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

go build -v -o /dev/null ./...

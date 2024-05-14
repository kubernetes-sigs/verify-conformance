#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

find . -name "*.go" | grep -E -v vendor | xargs gofmt -s -l -d -w -s

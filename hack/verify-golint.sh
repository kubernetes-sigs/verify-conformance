#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

go install golang.org/x/lint/golint@latest
find . -type f -name '*.go' -not -path './vendor/*' \
    | xargs -I{} echo golint -set_exit_status {}

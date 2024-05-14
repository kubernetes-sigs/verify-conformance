#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

go install golang.org/x/tools/cmd/goimports@latest
find . -type f -name '*.go' -not -path './vendor/*' | xargs -I{} goimports -w {}
if git diff --name-only --diff-filter=ACMRT | grep -E '(.*).go$'; then
  echo "error: changes detected, run 'find . -type f -name '*.go' -not -path './vendor/*' | xargs -I{} goimports -w {}'"
  exit 1
fi

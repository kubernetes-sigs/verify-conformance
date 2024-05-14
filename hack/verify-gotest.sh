#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

TMPDIR="$(mktemp -d)"
COVERPROFILEOUT="$TMPDIR/verify-conformance.out"
HTMLOUT="$TMPDIR/verify-conformance.html"
go test -cover -coverprofile="$COVERPROFILEOUT" -v ./...
go tool cover -html "$COVERPROFILEOUT" -o "$HTMLOUT"
echo "wrote: $HTMLOUT ($COVERPROFILEOUT)"

#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

FAILED_SCRIPTS=""
for SCRIPT in $(find ./hack -maxdepth 1 -mindepth 1 -iname 'verify-*' -not -iname 'verify-all*'); do
    FAILED=false
    echo "Running: $SCRIPT\n"
    sh -c "$SCRIPT" || FAILED=true
    if [ "$FAILED" = true ]; then
        FAILED_SCRIPTS="$SCRIPT $FAILED_SCRIPTS"
    fi
done

if [ -n "$FAILED_SCRIPTS" ]; then
    echo "error: there were one or more failures: $FAILED_SCRIPTS" >/dev/stderr
    exit 1
fi

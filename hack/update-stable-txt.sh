#!/bin/bash
# Prepare conformance metadata

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

curl -sSL https://storage.googleapis.com/kubernetes-release/release/stable.txt | tee ./kodata/metadata/stable.txt

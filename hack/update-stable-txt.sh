#!/bin/bash
# Prepare conformance metadata

cd "$(git rev-parse --show-toplevel)"

curl -sSL https://storage.googleapis.com/kubernetes-release/release/stable.txt | tee ./kodata/metadata/stable.txt

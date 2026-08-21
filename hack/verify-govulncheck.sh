#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)" || exit 1

go install golang.org/x/vuln/cmd/govulncheck@latest

# Ignore vulnerabilities that are known and accepted by the project.
export IGNORE="GO-2023-1901" # TektonCD issue without a fix yet!

govulncheck -format json ./... > vulns.json || true

REMAINING=$(jq -rs '
  [.[] | select(.finding != null)
       | .finding
       | select(.trace[0].function != null)
       | .osv]
  | unique
  | map(select(. as $id | ($ENV.IGNORE | split(" ") | index($id)) == null))
  | .[]
' vulns.json)

if [ -n "$REMAINING" ]; then
  echo "Unignored vulnerabilities found:"
  echo "$REMAINING"
  exit 1
fi
echo "Only ignored/known vulnerabilities present."

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

# Prepare conformance metadata

set -o errexit
set -o nounset
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

K8S_LATEST_VERSION=$(curl -L -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
K8S_LATEST_MINOR_VERSION="$(awk '{split($1,array, "."); print array[2]}' <<<$K8S_LATEST_VERSION)"
K8S_LAST_MINOR_VERSION=$(($K8S_LATEST_MINOR_VERSION - 2))
SETS=($(seq $K8S_LAST_MINOR_VERSION $K8S_LATEST_MINOR_VERSION))
rm -r ./kodata/conformance-testdata/

MANIFESTS=()
for SET in "${SETS[@]}"; do
  MANIFESTS+=("https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.$SET/test/conformance/testdata/conformance.yaml")
done

re="^.*([0-9].[0-9]{2}|master).*$"
for METADATA in ${MANIFESTS[*]}; do
  if [[ $METADATA =~ $re ]]; then
    version=${BASH_REMATCH[1]}
    echo "fetching for version '$version'"

    semver="v${version}"
    if [ "${version}" = master ]; then
      semver="${version}"
    fi
    output_file="./kodata/conformance-testdata/${semver}/conformance.yaml"
    mkdir -p ./kodata/conformance-testdata/$semver/
    curl -L \
      -s \
      -o "$output_file" \
      ${METADATA}
    ls $output_file
  else
    echo "warning: does not match on '$METADATA'"
  fi
done

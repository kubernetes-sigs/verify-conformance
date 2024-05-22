#!/bin/bash -x

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

cd "$(git rev-parse --show-toplevel)" || exit 0

C_DIR="/builds/$(basename $PWD)"
podman run --rm --network=host \
    -v "$PWD:$C_DIR:ro" --workdir "$C_DIR" \
    docker.io/golang:1.21.5-alpine3.18@sha256:9390a996e9f957842f07dff1e9661776702575dd888084e72d86eaa382ad56e3 \
      sh -c "
echo 'https://dl-cdn.alpinelinux.org/alpine/edge/testing' | tee -a /etc/apk/repositories ;
apk add --no-cache curl cosign ko git;
git config --global --add safe.directory $C_DIR ;
export KO_DOCKER_REPO=${KO_DOCKER_REPO:-localhost:5001/ghs}
./hack/publish.sh ${*:-}
"

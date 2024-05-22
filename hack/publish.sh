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

KO_FLAGS=""
# NOTE budget /bin/sh way
if echo "${@:-}" | grep -q '\-\-debug'; then
    set -x
    KO_FLAGS="--verbose $KO_FLAGS"
fi
if echo "${@:-}" | grep -q '\-\-sign'; then
    SIGN=true
fi
if echo "${@:-}" | grep -q '\-\-insecure'; then
    KO_FLAGS="--insecure-registry $KO_FLAGS"
fi

cd "$(git rev-parse --show-toplevel)" || exit 1

KO_DOCKER_REPO="${KO_DOCKER_REPO:-ghcr.io/cncf-infra/verify-conformance}"
APP_BUILD_HASH="${APP_BUILD_HASH:-$(git rev-parse HEAD | cut -c -8)}"
APP_BUILD_DATE="$(git show -s --format=%cd --date=format:'%Y.%m.%d.%H%M')"
APP_BUILD_VERSION="${APP_BUILD_VERSION:-0.0.0}"
APP_BUILD_MODE="${APP_BUILD_MODE:-development}"
TAG_FROM_COMMIT="$(git show -s --format=%cd --date=format:'%s')-$(git rev-parse HEAD | head -c8)"
IMAGE_DESTINATIONS="latest,$TAG_FROM_COMMIT"
if [[ -n "${CI_COMMIT_TAG:-}" ]]; then
  APP_BUILD_VERSION="${CI_COMMIT_TAG:-}"
  APP_BUILD_MODE=production
  IMAGE_DESTINATIONS="$APP_BUILD_VERSION,release-$TAG_FROM_COMMIT"
fi
echo "Commit made on '${APP_BUILD_DATE:-}'"

export KO_DOCKER_REPO \
    APP_BUILD_HASH \
    APP_BUILD_DATE \
    APP_BUILD_MODE \
    APP_BUILD_VERSION \
    IMAGE_DESTINATIONS

IMAGE="$(ko publish \
    --bare \
    --tags "${IMAGE_DESTINATIONS}" \
    $KO_FLAGS \
    .)"

if [ "${SIGN:-}" = true ]; then
    cosign sign --recursive -y "$IMAGE"
    cosign download sbom "$IMAGE" > /tmp/sbom-spdx.json
    cosign attest -y --recursive --predicate /tmp/sbom-spdx.json "$IMAGE"
fi

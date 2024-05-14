#!/bin/bash -x

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

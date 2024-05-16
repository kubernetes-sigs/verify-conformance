#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

cd "$(git rev-parse --show-toplevel)" || exit 1

CTR=docker
case "${1:-}" in
    podman)
        CTR=podman
        KIND_EXPERIMENTAL_PROVIDER=podman
        export KIND_EXPERIMENTAL_PROVIDER
        ;;
esac

# NOTE
# - from https://kind.sigs.k8s.io/docs/user/local-registry/

REG_NAME='kind-registry'
REG_PORT='5001' # NOTE important as on macOS: AirPlay server runs on port 5000
if [ "$("$CTR" inspect -f '{{.State.Running}}' "${REG_NAME}" 2>/dev/null || true)" != 'true' ]; then
  "$CTR" run \
    -d --restart=always -p "127.0.0.1:${REG_PORT}:5000" --network bridge --name "${REG_NAME}" \
    registry:2
fi

kind create cluster --config ./hack/local-dev/kind-config.yaml

REGISTRY_DIR="/etc/containerd/certs.d/localhost:${REG_PORT}"
for node in $(kind get nodes); do
  "$CTR" exec "${node}" mkdir -p "${REGISTRY_DIR}"
  cat <<EOF | "$CTR" exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR}/hosts.toml"
[host."http://${REG_NAME}:5000"]
EOF
done

if [ "$("$CTR" inspect -f='{{json .NetworkSettings.Networks.kind}}' "${REG_NAME}")" = 'null' ]; then
  "$CTR" network connect "kind" "${REG_NAME}"
fi

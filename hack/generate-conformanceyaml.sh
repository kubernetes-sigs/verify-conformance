#!/bin/bash
# Prepare conformance metadata

set -x

cd "$(git rev-parse --show-toplevel)"

K8S_LATEST_VERSION=$(curl -L -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
K8S_LATEST_MINOR_VERSION="$(awk '{split($1,array, "."); print array[2]}' <<<$K8S_LATEST_VERSION)"
SETS=($(seq 19 $K8S_LATEST_MINOR_VERSION))
rm -r ./kodata/conformance-testdata/

MANIFESTS=(
  https://raw.githubusercontent.com/cncf-infra/prow-config/master/docs/conformance_v1.{15..17}.yaml
  https://raw.githubusercontent.com/cncf-infra/prow-config/master/tests/conformance-1.18.yaml
  https://raw.githubusercontent.com/kubernetes/kubernetes/master/test/conformance/testdata/conformance.yaml
)
for SET in "${SETS[@]}"; do
  MANIFESTS+=("https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.$SET/test/conformance/testdata/conformance.yaml")
done

re="^.*([0-9].[0-9]{2}|master).*$"
for METADATA in ${MANIFESTS[*]}; do
  if [[ $METADATA =~ $re ]]; then
    version=${BASH_REMATCH[1]}

    semver="v${version}"
    if [ "${version}" = master ]; then
      semver="${version}"
    fi
    mkdir -p ./kodata/conformance-testdata/$semver/
    curl -L \
      -s \
      -o ./kodata/conformance-testdata/${semver}/conformance.yaml \
      ${METADATA}
  fi
done

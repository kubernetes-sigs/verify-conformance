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

cd "$(git rev-parse --show-toplevel)"

ACTIONS=$(
    for WORKFLOW in $(find .github/workflows -type f -name '*.yml' | sort | uniq); do
        ACTIONS=$(yq <$WORKFLOW e '.jobs.*.steps[].uses as $jobsteps | .jobs.*.uses as $jobuses | $jobsteps | [., $jobuses]' -o json |
            jq -rcMs --arg file "$WORKFLOW" '{"actions": . | flatten} | .file = $file')
        [ -z "${ACTIONS}" ] && continue
        echo -e "${ACTIONS}"
    done | jq -sc '.'
)
REPOSITORY="$(gh api repos/{owner}/{repo} --jq .full_name)"

for LINE in $(echo "$ACTIONS" | jq --arg REPOSITORY "$REPOSITORY" -rcM '.[] | .file as $file | .actions[] | . as $action_in_workflow | split("@") | .[0] as $action | $action | split("/") | .[0] as $org | .[1] as $repo | {"file": $file, "action": $action, "org": $org, "repo": $repo, "action_in_workflow": $action_in_workflow} | select(.action | contains($REPOSITORY) == false)'); do
    file="$(echo $LINE | jq -rcM .file)"
    action="$(echo $LINE | jq -rcM .action)"
    org="$(echo $LINE | jq -rcM .org)"
    repo="$(echo $LINE | jq -rcM .repo)"
    action_in_workflow="$(echo $LINE | jq -rcM .action_in_workflow)"

    echo "$file: $action; $org/$repo"

    default_branch="$(gh api repos/$org/$repo --jq .default_branch)"
    latest_commit_hash="$(gh api repos/$org/$repo/commits/$default_branch --jq '.sha')"
    latest_release_tag_name="$(gh api repos/$org/$repo/releases/latest --jq '.tag_name' 2>/dev/null || echo '')"
    if [ "$(echo $latest_release_tag_name | jq -r .message)" = "Not Found" ]; then
        latest_release_tag_name=false
    fi
    latest_release_commit_hash="$(gh api repos/$org/$repo/git/ref/tags/$latest_release_tag_name --jq .object.sha 2>/dev/null || echo '')"
    if [ "$(echo $latest_release_commit_hash | jq -r .message)" = "Not Found" ]; then
        latest_release_commit_hash=false
    fi

    commit_hash=
    if [ -n "$latest_release_commit_hash" ] && [ ! "$latest_release_commit_hash" = false ]; then
        commit_hash="$latest_release_commit_hash"
    else
        commit_hash="$latest_commit_hash"
    fi

    printf "$file: $action@$commit_hash (from $action_in_workflow)"
    if [ ! "$latest_release_tag_name" = false ]; then
        echo " # $latest_release_tag_name"
    else
        latest_release_tag_name="$default_branch"
    fi

    export FROM="$action_in_workflow" TO="$action@$commit_hash # $latest_release_tag_name"
    yq e -i 'with(.jobs.*.steps[] | select(.uses == env(FROM)); .uses = env(TO)) | with(.jobs.* | select(.steps | length == 0); del .steps)' "$file"
done

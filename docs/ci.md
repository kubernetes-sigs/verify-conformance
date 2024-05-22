# CI jobs

> GitHub Actions CI jobs and their definitions

## Code quality check

File: [code-quality-check.yml](../.github/workflows/code-quality-check.yml)

On: pull request

Group of parallel jobs to run on PR to ensure the code is to a high standard.

### gobuild

Calls: [hack/verify-gobuild.sh](../hack/verify-gobuild.sh).

Ensure the project compiles.

### gofmt

Calls: [hack/verify-gofmt.sh](../hack/verify-gofmt.sh).

Ensure the project is formatted correctly.

### goimports

Calls: [hack/verify-goimports.sh](../hack/verify-goimports.sh).

Ensure that the imports are formatted correctly.

### golangci-lint

Calls: [hack/verify-golangci-lint.sh](../hack/verify-golangci-lint.sh).

Ensure lint succeeds according to large standard set of rules.

### golint

Calls: [hack/verify-golint.sh](../hack/verify-golint.sh).

Ensure lint succeeds according to Go set of rules.

### gotest

Calls: [hack/verify-gotest.sh](../hack/verify-gotest.sh).

Ensure tests pass.

### govet

Calls: [hack/verify-govet.sh](../hack/verify-govet.sh).

Ensure no static analysis failures.

### govulncheck

Calls: [hack/verify-govulncheck.sh](../hack/verify-govulncheck.sh).

Ensure no known vulnerabilities.

### spellcheck

Calls: [hack/verify-spellcheck.sh](../hack/verify-spellcheck.sh).

Ensure no spelling mistakes.

## Release

File: [release.yml](../.github/workflows/release.yml)

On: release

Calls: [hack/publish.sh](../hack/publish.sh).

Builds and pushes a container image with Ko.

## Update conformance YAML

File: [update-conformance-yaml.yml](../.github/workflows/update-conformance-yaml.yml)

On: schedule (hourly)

Calls: [hack/generate-conformanceyaml.sh](../hack/generate-conformanceyaml.sh).

Iterates through list of Kubernetes versions and generates a folder containing conformance.yaml files to be consumed by the bot. The conformance.yaml files describe the tests required for conformance in the given release.

## Update Go Version

File: [update-go-version.yml](../.github/workflows/update-go-version.yml)

On: schedule (monthly)

A quality-of-life job to update the version in go.mod to the latest.

## Update stable txt

File: [update-stable-txt.yml](../.github/workflows/update-stable-txt.yml)

On: schedule (hourly)

Calls: [hack/update-stable-txt.sh](../hack/update-stable-txt.sh).

Cache the latest stable.txt file, containing the current latest stable version of Kubernetes.


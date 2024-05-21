# Resources

> Resources for Kubernetes conformance infrastructure

## Infra

The AWS account which is used is _cncf-apisnoop / 928655657136_.

The Prow cluster config is located in [_github.com/cncf-infra/prow-config_ repo under _infra/aws_](https://github.com/cncf-infra/prow-config/tree/master/infra/aws), it contains
- terraform cluster definition for EKS
  - state stored in S3 (_arn:aws:s3:::prow-cncf-io-tfstate_)
- prow configuration
- partial verify-conformance configuration (deprecated)

Configuration for IAM in the accounts, including the role which is used for continuous deployment is found in [github.com/cncf-infra/aws-infra/terraform/iam/main.tf](https://github.com/cncf-infra/aws-infra/blob/main/terraform/iam/main.tf).

There is a cluster in _cncf-apisnoop_ AWS account called _prow-cncf-io-eks_ (_arn:aws:eks:ap-southeast-2:928655657136:cluster/prow-cncf-io-eks_).

## Software

### verify-conformance

Repo located at [github.com/cncf-infra/verify-conformance](https://github.com/cncf-infra/verify-conformance)

Images are built using [`ko`](https://ko.build) and are published to [ghcr.io/cncf-infra/verify-conformance/verify-conformance](https://github.com/cncf-infra/verify-conformance/pkgs/container/verify-conformance%2Fverify-conformance), this is automated through the [release.yml](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/release.yml) workflow.

The bot has several metadata files baked into the container image which are updated by automations; these are
- [metadata/stable.txt](https://github.com/cncf-infra/verify-conformance/tree/main/kodata/metadata) :: for the latest known version of Kubernetes
  - updated by the [`hack/update-stable-txt.sh` script](https://github.com/cncf-infra/verify-conformance/blob/main/hack/update-stable-txt.sh)
- [conformance-testdata](https://github.com/cncf-infra/verify-conformance/tree/main/kodata/conformance-testdata) :: for the list of conformance e2e tests for each version
  - updated by the [`hack/generate-conformanceyaml.sh` script](https://github.com/cncf-infra/verify-conformance/blob/main/hack/generate-conformanceyaml.sh) which regenerates the folder entirely each time
  
The bot relies on a copy of godog's internal types for parsing the test results in the buffer, this copy is located in [internal/types/types.go](https://github.com/cncf-infra/verify-conformance/blob/main/internal/types/types.go) and the original is located at [github.com/cucumber/godog/internal/formatters/fmt_cucumber.go](https://github.com/cucumber/godog/blob/7f75c5d4ee9cd2e9d86b7ff62ebf38b9172d2c88/internal/formatters/fmt_cucumber.go#L116); these may change in the future.

#### Automations

workflows:

- [`go-build` (smoke test / does it compile?)](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/build.yml)
- [code quality check](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/code-quality-check.yml)
  - [`gofmt`](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/reusable-gofmt.yml)
  - [`golangci-lint`](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/reusable-golangci-lint.yml)
  - [`go test`](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/reusable-go-test.yml)
  - [`go vet`](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/reusable-go-vet.yml)
- [update conformance.yaml](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/update-conformance-yaml.yml)
- [update go version](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/update-go-version.yml)
- [update stable.txt](https://github.com/cncf-infra/verify-conformance/blob/main/.github/workflows/update-stable-txt.yml)

dependabot also is enabled, to update Go packages and GitHub Actions versions.

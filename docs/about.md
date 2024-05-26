# About

> the way verify-conformance operates and more

## Operation

The bot uses a GitHub App or auth token to authenticate as an app or user respectively.

It performs a GitHub search query on the configured repo with the flag `--repo` looks for open PRs. The results are then determined to be conformance submission PRs or not -- this is important as to not run on changes like random documentation updates. Polling and webhooks are used to ensure PRs get checked. These functions take place in [internal/plugin/plugin.go](../internal/plugin/plugin.go).

The bot uses the cucumber format for writing test directives in a human (usually English) readable format. It must find where the files are located, under kodata directory and features. Take the following scenario where the directives `the files in the PR` and `the files included in the PR are only:` both map to Go functions in [internal/suite/suite.go](../internal/suite/suite.go).

```feature
Feature: verify conformance product submission PR
...

  Scenario: submission only contains required files
    Given the files in the PR
    Then the files included in the PR are only: "README.md, PRODUCT.yaml, e2e.log, junit_01.xml"

...
```

Godog is the test framework implementation for Cucumber in Go. It can output the results in several formats. Unfortunately, the Godog package doesn't export it's *Feature* types, so they are vendored in over at [internal/types/types.go](../internal/types/types.go). This is useful so the raw JSON bytes can be structured and parsed to be processed into data the bot will output like comments, labels and status.

The bot configures a suite run of these tests from the feature file, feeding in the PR. Several bits of data are collected for the test run, like: labels, changes in PR, *PRODUCT.yaml* URL data (logo datatypes etc...), [cached](../kodata/metadata/stable.txt) Kubernetes [stable.txt](https://dl.k8s.io/release/stable.txt). The testsuite is then run and the results of comment, labels and state are used to reconcile then comments, labels and status.

The test suite consists of

- tests passing and present in *junit_01.xml*
- PR submission up to standard (ease of bot understanding)
- files are valid

for a more detailed look, see [kodata/features/verify-conformance.feature](../kodata/features/verify-conformance.feature).

The required tests are described in conformance.yaml files cached in [kodata/conformance-testdata/](../kodata/conformance-testdata/) and under the specific version, these files come from [git.k8s.io/kubernetes/test/conformance/testdata/conformance.yaml](https://git.k8s.io/kubernetes/test/conformance/testdata/conformance.yaml).

Cucumber was chosen to provide better insight to all for what is required for conformance, making describing the behaviour apart of implementing a test via Test Driven Development (TDD).

## Notes

- built with [`ko`](https://ko.build) and reads files from `kodata/` locally, then accessible in `/var/run/ko` when running as a container


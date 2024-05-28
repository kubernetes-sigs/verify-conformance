# Maintenance

> Common assortment of bot maintenance tasks

## Generating conformance.yaml metadata

The [**kodata/conformance-testdata**](../kodata/conformance-testdata) folder is a managed folder, where the script [**hack/generate-conformanceyaml.sh**](../hack/generate-conformanceyaml.sh) recreates the sub-folders according to the last three release versions of Kubernetes and fetching the respective conformance.yaml files.

This process is automated due to a GitHub Action workflow, called [update-conformance-yaml.yml](../.github/workflows/update-conformance-yaml.yml), where PRs are automatically generated and merged.

The structure will end up like

- *kodata/*
  - *conformance-testdata/*
    - *v1.28/*
      - *conformance.yaml*
    - *v1.29/*
      - *conformance.yaml*
    - *v1.30/*
      - *conformance.yaml*

## Updating stable.txt

The [kodata/metadata/stable.txt](../kodata/metadata/stable.txt) file is a managed file which contains a string of the latest version known of Kubernetes. This file is synced from [dl.k8s.io/release/stable.txt](https://dl.k8s.io/release/stable.txt) and is cached to slightly reduce the amount of requests outbound.

This process is automated due to a GitHub Action workflow, called [update-stable-txt.yml](../.github/workflows/update-stable-txt.yml), where PRs are automatically generated and merged.

## Adding new confomance results checks

First, the idea must be modeled in [verify-conformance.feature](../kodata/feature/verify-conformance.feature). Create a new scenario like

```feature
  Scenario: submission is only one product
    the submission seems to contain files of multiple Kubernetes release versions or products. Each Kubernetes release version and products should be submitted in a separate PRs

    Given the files in the PR
    Then there is only one path of folders
```

Or for an outline

```feature
  Scenario: [TITLE]
    [ERROR MESSAGE RESPONSE]
  
  Given [FIRST STATEMENT]
  And [SECOND STATEMENT]
  Then [THIRD STATEMENT]
```

*Important notes*:

- use declarative phrasing for all text. Such as: _there is only one path of folders_
- read up on the syntax at [cucumber.io/docs/gherkin/reference/](https://cucumber.io/docs/gherkin/reference/)

Next, connect the phrase to a function in [suite.go](../internal/suite/suite.go). Inside of _InitializeScenario_, create a step with a regexp valid to connect it. Such as:

```go
...
	ctx.Step(`^the files in the PR`, s.theFilesInThePR)
...

func (s *PRSuite) theFilesInThePR() error {
...
}
```

Or if a value must be captured and passed to the function, a statement like:

```go
...
	ctx.Step(`^the title of the PR matches "([^"]*)"$`, s.theTitleOfThePRMatches)
...

func (s *PRSuite) theTitleOfThePRMatches(match string) error {
...
}
```

After completing those steps the testcase will then be connected.

## Adding new code tests

Install gotests

```go
go install github.com/cweill/gotests@latest
```

Generate new boilerplate

```go
gotests -w -all ./internal/
```

Tests will be generated looking something like the following

```go
func Test_FUNC_NAME_HERE(t *testing.T) {
	tests := []struct {
		name string
		want options
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FUNC_NAME_HERE(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FUNC_NAME_HERE() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

The tests are testcase driven and are generally declarative. Take a look at the other tests in this project or any other update-to-date project using standard Go tests.

## Releasing

Versions are tagged in the format of YYYY-mm-dd-HHMM or 2024-05-23-1452 for example.

1. Navigate to https://github.com/cncf-infra/verify-conformance/releases/new
2. Under choose a tag input a version in the specified format above
3. Click _Generate release notes_
4. Click _Publish release_

A new container is now built and tagged to the version specified, like _ghcr.io/cncf-infra/verify-conformance:VERSION_.

## Deployment

The verify-conformance bot is deployed in GitHub Actions.

First, navigate to https://github.com/organizations/cncf-infra/settings/apps/prow-cncf-io (NOTE: will change). Note down the App ID and download a new private key.

Next, navigate to https://github.com/cncf/k8s-conformance/settings/secrets/actions and create a repository secret called `GH_APP_ID` with the value of the App ID copied earlier. Create a second repository secret, this one called `GH_APP_PRIVATE_KEY` with the value of the private key but base64 encoded. Create a final repository secret called `GH_APP_HMAC` and set it to a value matching _Webhook secret_ on the GitHub App page.

The workflow can be deployed in a pipeline on a cronjob with a call to run the container image with specific flags. Like so:

```yaml
mkdir -p ./tmp/
echo "$GH_APP_PRIVATE_KEY" | base64 -d > ./tmp/github-app-private-key
echo "$GH_APP_HMAC" > ./tmp/hmac

docker run --rm \
  -v "$PWD:$PWD:ro" \
  --workdir "$PWD" \
  ghcr.io/cncf-infra/verify-conformance:latest \
    --github-endpoint=https://api.github.com \
    --dry-run=false \
    --github-app-id="$GH_APP_ID" \
    --github-app-private-key-path="$PWD/tmp/github-app-private-key" \
    --hmac-secret-file=$PWD/tmp/hmac \
    --repo="$REPO"
```

see: https://github.com/cncf/k8s-conformance/blob/master/.github/workflows/verify-conformance.yml

### GitHub App

In the case a new GitHub App needs to be set up, navigate to a page like https://github.com/organizations/cncf-infra/settings/apps/new and fill in the values like

| Field                                  | Value                                                                                                               |
|----------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| GitHub App name                        | Kubernetes Conformance bot                                                                                          |
| Description                            | _verifying conformance product submissions_                                                                         |
| Homepage URL                           | https://github.com/kubernetes-sigs/verify-conformance                                                               |
| Webhook -> Active                      | false (not set)                                                                                                     |
| Permissions -> Repository permissions  | Commit Statuses : Read and write, Contents : Read and write, Issues : Read and write, Pull requests: Read and write |
| Where can this GitHub App be installed | Any account (for testing)                                                                                           |

Next, go back up to deployment in the header above to deploy this with the application.

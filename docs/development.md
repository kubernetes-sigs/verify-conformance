# Development

> Set up a local environment for development

## Notes

- the bot will log in and make comments are your GitHub user or whatever user the token belongs to

## Environment

install tools (macOS or Linux)

- [go](https://go.dev)
- [gh](https://cli.github.com/)

```
brew install go gh
```

(NOTE: **example**)

log in to GitHub with `gh`

```
gh auth login
```

write secrets

```
cd "$(git rev-parse --show-toplevel)"
mkdir -p ./hack/local-dev/tmp/
echo "$(openssl rand -base64 15)" > ./hack/local-dev/tmp/hmac
gh auth token > ./hack/local-dev/tmp/token
```
(**NOTE**: avoid committing these values)

start up ghproxy

```sh
docker run \
  -d \
  -p 8888:8888 \
  --name ghproxy \
  gcr.io/k8s-prow/ghproxy:v20240723-dbbd2d86b
```

# Development loop

run locally

```sh
go run . \
  --github-endpoint=http://localhost:8888 \
  --github-endpoint=https://api.github.com \
  --dry-run=false \
  --hmac-secret-file=./hack/local-dev/tmp/hmac \
  --github-token-path=./hack/local-dev/tmp/token \
  --repo=cncf-infra/k8s-conformance
```


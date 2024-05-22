# Development

> Set up a local environment for development

# Notes

- currently pushes to a public container registry
- the bot will log in and make comments are your GitHub user or whatever user the token belongs to

# Environment

install tools (macOS or Linux)

- [podman](https://podman.io) or [docker](https://docker.com)
- [ko](https://ko.build)
- [kind](https://kind.sigs.k8s.io)
- [kustomize](https://kustomize.io)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [go](https://go.dev)
- [gh](https://cli.github.com/)

```
brew install podman ko kind kustomize kubectl go gh
```

(NOTE: **example**)

log in to GitHub with `gh` and package write permissions

```
gh auth login -s write:packages
```

log into ghcr.io (**optional**)

```
gh auth token | ko login ghcr.io --username "$(gh api user --jq .login)" --password-stdin
```

write secrets (**example**)

```
cd "$(git rev-parse --show-toplevel)"
mkdir -p ./hack/local-dev/tmp/
echo "$(openssl rand -base64 15)" > ./hack/local-dev/tmp/hmac
gh auth token > ./hack/local-dev/tmp/token
```

(**NOTE**: avoid committing these values)

create a cluster

```
./hack/local-dev/start-kind.sh
```

# Development loop

build image

```
export KO_DOCKER_REPO=localhost:5001/verify-conformance
IMAGE="$(ko build --base-import-paths .)"
```

(**NOTE**: feel free to swap out registry above)

configure components (**optional**)

```
cd ./hack/local-dev/
kustomize edit set image ko://sigs.k8s.io/verify-conformance="$IMAGE"
```

(**NOTE**: avoid committing this change)

apply

```
cd "$(git rev-parse --show-toplevel)"
kustomize build ./hack/local-dev/ | kubectl apply -f -
```

observe resources

```
kubectl -n prow get all
```

# Clean up environment

teardown

```
kind delete cluster
docker rm -f kind-registry
```

# Tips

read the logs

```
kubectl -n prow logs -l app=verify-conformance --tail=50 -f
```

restart it

```
kubectl -n prow rollout restart deployment verify-conformance
```

compile test

```
go build -o bin/ .
```


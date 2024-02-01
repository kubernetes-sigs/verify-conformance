# Building

Locally with Go

```shell
go build -o bin/ .
```

Locally with Ko

```shell
export KO_DOCKER_REPO=localhost:5001/verify-conformance
ko build --base-import-paths .
```

With cloudbuild

```shell
REGION=australia-southeast1

gcloud builds submit \
  --region=$REGION \
  --config=cloudbuild.yaml \
  .
```

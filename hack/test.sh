#!/usr/bin/env sh

go test -v -cover -coverprofile=/tmp/verify-conformance.out ./...
go tool cover -html /tmp/verify-conformance.out -o /tmp/verify-conformance.html
open /tmp/verify-conformance.html

FROM golang:1.13.4
COPY ./ /tmp/build
WORKDIR /tmp/build
RUN go get && go build . && mkdir -p /plugin && cp verify-conformance-release /plugin
WORKDIR /plugin
COPY verify-conformance-release .
ENTRYPOINT ["/plugin/verify-conformance-release"]

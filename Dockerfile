FROM golang:1.13.4
COPY ./ /tmp/build
WORKDIR /tmp/build
RUN go get && go build . && mkdir -p /plugin && cp verify-conformance-tests /plugin
WORKDIR /plugin
COPY verify-conformance-request .
ENTRYPOINT ["/plugin/verify-conformance-request"]

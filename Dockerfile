FROM golang:1.16.0-alpine3.13 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH="" go build \
  -a \
  -installsuffix cgo \
  -ldflags "-extldflags '-static' -s -w" \
  -o bin/verify-conformance-release \
  main.go

FROM alpine:3.11 as extras
RUN apk add tzdata ca-certificates
RUN adduser -D user

FROM scratch AS final
WORKDIR /app
ENV PATH=/app/bin
COPY --from=build /app/bin/verify-conformance-release /app/bin/verify-conformance-release
COPY --from=extras /etc/passwd /etc/passwd
COPY --from=extras /etc/group /etc/group
COPY --from=extras /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=extras /etc/ssl /etc/ssl
EXPOSE 8888
USER user
ENTRYPOINT ["/app/bin/verify-conformance-release"]

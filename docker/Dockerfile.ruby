FROM golang:1.9-alpine AS build
ARG VERSION=0.0.0
ARG GIT_HASH=c0ff33
WORKDIR /go/src/github.com/target/flottbot/
RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
COPY / .
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-X github.com/target/flottbot/version.Version=${VERSION} -X github.com/target/flottbot/version.GitHash=${GIT_HASH}" \
    -o flottbot .

FROM ruby:2.4.3-alpine3.7
RUN apk add --no-cache ruby-dev build-base
RUN mkdir config
COPY --from=build /go/src/github.com/target/flottbot/flottbot .
EXPOSE 8080 80

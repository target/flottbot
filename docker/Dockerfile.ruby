FROM golang:1.14-alpine AS build
ARG VERSION
ARG GIT_HASH
ENV GO111MODULE=on

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION} -X github.com/target/flottbot/version.GitHash=${GIT_HASH}" \
    -o flottbot ./cmd/flottbot

FROM ruby:2.7-alpine
ENV USERNAME=flottbot
ENV UID=900
RUN apk add --no-cache ca-certificates ruby-dev build-base && mkdir config && adduser -S -u "$UID" "$USERNAME"
COPY --from=build /src/flottbot /flottbot

EXPOSE 8080 3000 4000

USER ${USERNAME}
CMD ["/flottbot"]

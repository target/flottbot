FROM docker.io/golang:1.19.5-alpine@sha256:2381c1e5f8350a901597d633b2e517775eeac7a6682be39225a93b22cfd0f8bb AS build
ARG VERSION

# needed for vcs feature introduced in go 1.18+
RUN apk add --no-cache git

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION}" \
  -o flottbot ./cmd/flottbot

FROM docker.io/ruby:3.2.0-alpine@sha256:c690d3b45ef65e0098104c272f1736b0595b824908d5639c3a3d17636581a905

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk add --no-cache ca-certificates ruby-dev build-base && mkdir config &&  \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 8080 3000 4000

USER ${USERNAME}
CMD ["/flottbot"]

FROM docker.io/golang:1.18-alpine@sha256:dda10a0c69473a595ab11ed3f8305bf4d38e0436b80e1462fb22c9d8a1c1e808 AS build
ARG VERSION

# needed for vcs feature introduced in go 1.18
RUN apk add --no-cache git

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION}" \
  -o flottbot ./cmd/flottbot

FROM docker.io/ruby:3.1-alpine@sha256:ed6cafb49b8c5893f3d7d32a8c27ea39ec67520d2684487c4c70fa618f33a76f

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

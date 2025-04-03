FROM --platform=${BUILDPLATFORM} docker.io/golang:1.24.2-alpine@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

# needed for vcs feature introduced in go 1.18+
RUN apk add --no-cache git

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION}" \
  -o flottbot ./cmd/flottbot

FROM docker.io/ruby:3.4.2-alpine@sha256:cb6a5cb7303314946b75fa64c96d8116f838b8495ffa161610bd6aaaf9a70121

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk add --no-cache ca-certificates curl jq ruby-dev build-base && mkdir config &&  \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 8080 3000 4000

USER ${USERNAME}
CMD ["/flottbot"]

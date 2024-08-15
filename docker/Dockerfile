FROM --platform=${BUILDPLATFORM} docker.io/golang:1.23.0-alpine@sha256:d0b31558e6b3e4cc59f6011d79905835108c919143ebecc58f35965bf79948f4 AS build

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

FROM docker.io/alpine:3.20.2@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk --no-cache add ca-certificates curl jq && mkdir config && \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 3000 4000 8080

USER ${USERNAME}
CMD [ "/flottbot" ]

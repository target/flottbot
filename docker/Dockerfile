FROM docker.io/golang:1.19.3-alpine@sha256:dc4f4756a4fb91b6f496a958e11e00c0621130c8dfbb31ac0737b0229ad6ad9c AS build
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

FROM docker.io/alpine:3.16.3@sha256:b95359c2505145f16c6aa384f9cc74eeff78eb36d308ca4fd902eeeb0a0b161b

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk --no-cache add ca-certificates && mkdir config && \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 3000 4000 8080

USER ${USERNAME}
CMD [ "/flottbot" ]

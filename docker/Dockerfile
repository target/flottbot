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

FROM alpine:3.11
ENV USERNAME=flottbot
ENV UID=900
RUN apk --no-cache add ca-certificates && mkdir config && adduser -S --u "$UID" "$USERNAME"
COPY --from=build /src/flottbot /flottbot

EXPOSE 3000 4000 8080

USER ${USERNAME}
CMD [ "/flottbot" ]
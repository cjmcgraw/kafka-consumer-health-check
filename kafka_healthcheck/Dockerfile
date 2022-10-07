FROM golang:1.19-bullseye

WORKDIR /usr/src/app
ENV GOCACHE /usr/src/app/.cache
RUN go env GOCACHE
COPY lib/ .
RUN --mount=type=cache,target=/usr/src/app/.cache go mod download && go mod verify
RUN --mount=type=cache,target=/usr/src/app/.cache go build -v -o healthcheck ./cmd/healthcheck/main.go
ENTRYPOINT ["./healthcheck"]


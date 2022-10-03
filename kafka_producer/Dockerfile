FROM golang:1.19-bullseye

WORKDIR /usr/src/app
COPY lib/go.* .
RUN go mod download && go mod verify
COPY lib/cmd/producer ./
RUN go build -v -o /usr/local/bin/app .
ENTRYPOINT ["app"]


FROM golang:1.19-bullseye

WORKDIR /usr/src/app
COPY lib .
RUN go mod download && go mod verify
RUN go build -v -o consumer cmd/consumer/main.go
ENTRYPOINT ["./consumer"]


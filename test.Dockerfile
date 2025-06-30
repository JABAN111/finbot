FROM golang:1.24

WORKDIR /builder

COPY go.mod go.sum /builder/
COPY . .



RUN go test -v ./...
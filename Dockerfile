FROM golang:1.24 AS builder

WORKDIR /builder

COPY go.mod go.sum /builder/
COPY . .

ENV CGO_ENABLED=0
RUN go build -o finbot

FROM alpine:3.20 AS runner

WORKDIR /runner

COPY --from=builder /builder/finbot finbot

ENTRYPOINT ["/runner/finbot"]
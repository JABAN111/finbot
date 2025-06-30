FROM golang:1.24

WORKDIR /builder

COPY go.mod go.sum /builder/
COPY . .



CMD ["go", "test", "-v", "./..."]
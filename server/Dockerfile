FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd

FROM ubuntu:24.04
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/cmd/docs ./cmd/docs

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

EXPOSE 8080
CMD ["./main"]

# Build stage
FROM golang:1.19 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/GoGameServer ./src/main

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/GoGameServer ./GoGameServer
COPY bin/config/config.toml ./config.toml
ENTRYPOINT ["./GoGameServer"]
CMD ["run","game","0"]

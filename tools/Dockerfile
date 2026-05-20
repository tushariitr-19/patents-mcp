FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o patents-mcp-server ./cmd/server/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/patents-mcp-server .
ENTRYPOINT ["./patents-mcp-server"]
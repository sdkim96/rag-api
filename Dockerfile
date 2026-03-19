FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /rag-api ./cmd/rag-api

FROM alpine:3.21

COPY --from=builder /rag-api /rag-api

EXPOSE 8080

ENTRYPOINT ["/rag-api", "-mode", "http", "-addr", ":8080"]

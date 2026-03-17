FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /dating ./cmd/dating

FROM alpine:3.21

COPY --from=builder /dating /dating

EXPOSE 8080

ENTRYPOINT ["/dating", "-mode", "http", "-addr", ":8080"]

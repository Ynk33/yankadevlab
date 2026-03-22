# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o auth .

# ---- Stage 2: Runtime ----
FROM alpine:3.21

COPY --from=builder /app/auth /auth
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

CMD ["/auth"]
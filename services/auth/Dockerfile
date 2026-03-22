# ---- Stage 1: Build ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o auth .

# ---- Stage 2: Runtime ----
FROM alpine:3.21

COPY --from=builder /app/auth /auth

EXPOSE 8080

CMD ["/auth"]
# ---- Stage 1: Build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o monitoring ./cmd

# ---- Stage 2: Runtime ----
FROM alpine:3.21

COPY --from=builder /app/monitoring /monitoring

EXPOSE 8080

CMD ["/monitoring"]

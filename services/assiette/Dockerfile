# ---- Stage 1: Build frontend ----
FROM node:22-alpine AS web

WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# ---- Stage 2: Build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o assiette . && CGO_ENABLED=0 go build -o seed ./cmd/seed

# ---- Stage 3: Runtime ----
FROM alpine:3.21

COPY --from=builder /app/assiette /assiette
COPY --from=builder /app/seed /seed
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

CMD ["/assiette"]

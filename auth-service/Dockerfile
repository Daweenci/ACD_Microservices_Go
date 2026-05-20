# ── Stage 1: Builder ─────────────────────────────────────────────────
# Kompiliert das Go-Binary (wie im Artikel mit dem Builder-Pattern)
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Cross-kompiliertes, statisch gelinktes Binary – kein CGO
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o auth-service ./cmd/main.go

# ── Stage 2: Minimaler Container (FROM scratch) ───────────────────────
# Enthält NUR das Binary – wie im Artikel beschrieben (~7 MB)
FROM scratch

COPY --from=builder /build/auth-service /app/auth-service

WORKDIR /app
EXPOSE 8081

ENTRYPOINT ["/app/auth-service"]

FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o shop-service ./cmd/main.go

FROM scratch

COPY --from=builder /build/shop-service /app/shop-service

WORKDIR /app
EXPOSE 8082

ENTRYPOINT ["/app/shop-service"]

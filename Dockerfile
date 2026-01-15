#building the binary
FROM golang:1.23.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GIN_MODE=release go build -o ./cmd/bin/app ./cmd/api/main.go

#copying binary into final image
FROM alpine:3.19

RUN apk add --no-cache tzdata curl

WORKDIR /app

COPY --from=builder /app/cmd/bin/app .
COPY ./static /app

HEALTHCHECK --interval=10s --timeout=2s --retries=5 \
  CMD curl -fs http://localhost:8080/health || exit 1

CMD ["./app"]
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY cmd/ ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /main ./cmd/api/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /

COPY --from=builder /main .

EXPOSE 8080

CMD ["./main"]
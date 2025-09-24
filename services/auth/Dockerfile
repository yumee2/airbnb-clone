FROM golang:1.24.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/auth .
COPY --from=builder /app/config  ./config


EXPOSE 8000

CMD ["./auth"]
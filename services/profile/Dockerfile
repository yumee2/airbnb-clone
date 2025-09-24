FROM golang:1.24.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o profile ./cmd

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/profile .
COPY --from=builder /app/config  ./config


EXPOSE 8002

CMD ["./profile"]
FROM golang:1.24.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o booking ./cmd

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/booking .
COPY --from=builder /app/config  ./config


EXPOSE 8004

CMD ["./booking"]
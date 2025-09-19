FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o my-service ./cmd/my-service/

EXPOSE 8080

CMD ["./my-service"]
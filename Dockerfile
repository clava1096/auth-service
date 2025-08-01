# syntax=docker/dockerfile:1
FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o auth-service

EXPOSE 8080

CMD ["./auth-service"]
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./ # No dependencies yet
RUN go mod download

COPY *.go ./

RUN go build -o /ping-pong

FROM alpine:3.18

WORKDIR /

COPY --from=builder /ping-pong /ping-pong

EXPOSE 8080

CMD ["/ping-pong"]

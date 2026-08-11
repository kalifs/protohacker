FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY main.go ./

RUN go build -o /prime ./main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /prime /usr/local/bin/prime

EXPOSE 8080

CMD ["/usr/local/bin/prime"]

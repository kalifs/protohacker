FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY main.go ./

RUN go build -o /chat ./main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /chat /usr/local/bin/chat

EXPOSE 8080

CMD ["/usr/local/bin/chat"]

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY main.go ./

RUN go build -o /db ./main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /db /usr/local/bin/db

EXPOSE 5000

CMD ["/usr/local/bin/db"]

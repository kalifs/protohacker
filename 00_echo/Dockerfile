FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY main.go ./

RUN go build -o /echo ./main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /echo /usr/local/bin/echo

EXPOSE 8080

CMD ["/usr/local/bin/echo"]

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY main.go ./

RUN go build -o /means ./main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /means /usr/local/bin/means

EXPOSE 8080

CMD ["/usr/local/bin/means"]

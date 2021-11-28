FROM golang:1.17.3-alpine3.14 AS builder
COPY . /app
WORKDIR /app
RUN go build -o http-server

FROM alpine:3.14
EXPOSE 8080
ENV VERSION=1.0
COPY --from=builder /app/http-server /http-server
ENTRYPOINT ["/http-server"]
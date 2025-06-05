FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o dns-server main.go
FROM alpine
WORKDIR /build
COPY --from=builder /build/dns-server /build/dns-server
CMD ["./dns-server","--server"]

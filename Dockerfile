# 1 - Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /services
COPY . .
RUN go mod download
RUN go build -o blob main.go

# 2 - Final stage
FROM alpine:latest
WORKDIR /services
COPY --from=builder /services/blob /services/blob
EXPOSE 3000
CMD ["/services/blob"]

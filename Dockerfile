# Builder stage
FROM golang:1.24-alpine3.20 AS builder
WORKDIR /go/src/warehouse-service
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o warehouse-service cmd/main.go

# Final image
FROM alpine:latest
EXPOSE 80

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /go/src/warehouse-service/warehouse-service .

CMD ["./warehouse-service"]

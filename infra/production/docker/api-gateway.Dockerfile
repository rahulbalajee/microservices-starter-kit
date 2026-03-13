FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/services/api-gateway/api-gateway /
ENTRYPOINT ["/api-gateway"]

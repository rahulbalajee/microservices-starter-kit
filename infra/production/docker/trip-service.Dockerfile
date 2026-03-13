FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/trip-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o trip-service ./cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/services/trip-service/trip-service /
ENTRYPOINT ["/trip-service"]

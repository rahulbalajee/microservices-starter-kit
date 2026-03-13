FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/driver-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o driver-service

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/services/driver-service/driver-service /
ENTRYPOINT ["/driver-service"]

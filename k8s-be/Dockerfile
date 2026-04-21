# Build the Binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./ 
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cloudcrush-api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o cloudcrush-worker ./cmd/worker/main.go

# Create final image
FROM alpine:latest AS final

WORKDIR /root

COPY --from=builder /app/cloudcrush-api .
COPY --from=builder /app/cloudcrush-worker .

EXPOSE 8080
CMD ["./cloudcrush-api"]






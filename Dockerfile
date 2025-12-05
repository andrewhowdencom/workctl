# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o workctl main.go

# Run stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /app/workctl /workctl

EXPOSE 8080

ENTRYPOINT ["/workctl", "serve"]

# Build stage
FROM golang:1.23.3 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

ARG GOOS=linux
ARG GOARCH=amd64
ARG BINARY_NAME=patron

RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o /output/${BINARY_NAME} .

# Final image
FROM alpine:latest

WORKDIR /
ARG BINARY_NAME=patron
COPY --from=builder /output/${BINARY_NAME} .

ENTRYPOINT ["/bin/true"]
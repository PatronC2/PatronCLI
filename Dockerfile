# Build stage
FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

ARG GOOS=linux
ARG GOARCH=amd64
ARG BINARY_NAME=patron
ARG TAG=snapshot
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
  go build -trimpath \
  -ldflags "-s -w \
    -X patroncli/version.Tag=${TAG} \
    -X patroncli/version.Commit=${COMMIT} \
    -X patroncli/version.Date=${BUILD_DATE}" \
  -o /out/${BINARY_NAME} .

# Final image
FROM alpine:latest

WORKDIR /
ARG BINARY_NAME=patron
COPY --from=builder /output/${BINARY_NAME} .

ENTRYPOINT ["/bin/true"]
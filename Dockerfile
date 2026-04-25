ARG GO_VERSION=1.25.6

FROM golang:${GO_VERSION} AS builder
WORKDIR /app

ARG GOOS=linux
ARG GOARCH=amd64
ARG BINARY_NAME=patron

ARG TAG=snapshot
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY go.mod ./
RUN go mod download
COPY . .

RUN mkdir -p /output && \
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -trimpath \
      -ldflags "-s -w \
        -X patroncli/version.Tag=${TAG} \
        -X patroncli/version.Commit=${COMMIT} \
        -X patroncli/version.Date=${BUILD_DATE}" \
      -o /output/${BINARY_NAME} .

# Final image
FROM scratch

WORKDIR /
ARG BINARY_NAME=patron
COPY --from=builder /output/${BINARY_NAME} .

CMD [ "/${BINARY_NAME}" ]

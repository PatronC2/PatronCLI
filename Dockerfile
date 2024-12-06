FROM golang:1.23.3 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

ARG GOOS=linux
ARG GOARCH=amd64

RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o /root/main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /root/main .

CMD ["./main"]

FROM golang:1.25 AS builder

RUN mkdir -p /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# ARG VERSION
# RUN make dep
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./main.go

FROM alpine:3.14
RUN apk update && apk add --no-cache ca-certificates

RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /app/server /app/server

ENTRYPOINT ["/app/server"]

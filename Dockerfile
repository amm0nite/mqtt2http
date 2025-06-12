# Build stage
FROM golang:1.24.3-alpine AS build

WORKDIR /workspace

RUN apk add --no-cache ca-certificates

ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app ./cmd/mqtt2http.go

# Run stage
FROM alpine

RUN apk add --no-cache ca-certificates

RUN adduser -D app

USER app
WORKDIR /home/app

COPY --from=build /workspace/app ./

CMD ["./app"]
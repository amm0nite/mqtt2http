FROM golang:1.20-bookworm as build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app ./cmd/mqtt2http.go

FROM debian:bookworm

RUN apt-get update && apt-get install -y ca-certificates

RUN useradd app
USER app

WORKDIR /home/app

COPY --from=build /workspace/app ./

CMD ["./app"]

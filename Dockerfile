FROM golang:1.18 as build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app ./cmd/mqtt2http/mqtt2http.go

FROM debian:bullseye

RUN useradd app
USER app

WORKDIR /home/app

COPY --from=build /workspace/app ./

CMD ["./app"]
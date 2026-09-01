FROM golang:1.27-alpine

WORKDIR /app

COPY go.mod .

COPY main.go .

RUN go build -o app .

CMD ["./app"]
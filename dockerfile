FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /api ./cmd/api
RUN go build -o /consumer ./cmd/consumer

EXPOSE 8080

CMD ["/api"]

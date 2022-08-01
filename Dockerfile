FROM golang:1.18.4-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY .env ./
COPY *.go ./

RUN go build -o /mux-api
EXPOSE 8001
CMD ["/mux-api"]
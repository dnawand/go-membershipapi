FROM golang:1.17.8-alpine3.15

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -v cmd/broker/main.go

EXPOSE 8080
EXPOSE 8081

ENTRYPOINT [ "./main" ]
FROM golang:1.17.8-alpine3.15 AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build cmd/api/main.go

FROM alpine:3.15.2

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
EXPOSE 8081
ENTRYPOINT [ "./main" ]
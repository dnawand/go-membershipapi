# Membership API

## Running

The API will run on `localhost:8080` and `localhost:8081` in order to serve the swagger file.

### Local

If you have go 1.17 installed, run the following in order:

1. `go mod tidy`
2. `go build cmd/api/main.go`
3. `./main`

You can also use `make runlocal`.

You can set `GIN_MODE` on the command with `GIN_MODE=release go run cmd/api/main.go`

By default, you can't pause subscription while in trial period. If you want to disable it
set the environment variable `ALLOW_PAUSE_ON_TRIAL` with any non-empty string locally.

### Docker

You can set `ALLOW_PAUSE_ON_TRIAL` with any non-empty string in `docker-compose` file before running it.

Since multiple services are describe on docker-file, especify the api one when running:
```bash
docker-compose up -d --build membershipapi
```

You can also use `make run`.

## Documentation

You can get the API documentation as swagger by two means:

1. inside the `cmd/api/swagger` there will be a `swagger.jon` file.
2. after running the API, access `localhost:8081/swagger/swagger/` on a browser.

Also, there are a postman collection and environment that you can import. It's in `collection` folder.

## Tests

Run manually `go test -v -cover -count=1 ./...` or use run `make test`.

## Improvement

Other componets could be added in general. Eg.: middleware for authentication, propagation of logging and context, 
more unity tests, etc.

# Membership API

## Running

The API will run on `localhost:8080` and `localhost:8081` in order to serve the swagger file.

### Local

With a postgres instance running, set this environment variables in order to allow the service to connect to database:
- `DB_HOST` - defaults to `localhost`
- `DB_PORT` - defaults to `5432`
- `DB_NAME` - defaults to `membership`
- `DB_USER` - defaults to `postgres`
- `DB_PW` - defaults to `secretpw`

Then run with `go build cmd/api/main.go`, and then `./main`.

You can also use `make runlocal`.

You can set `GIN_MODE` on the command with `GIN_MODE=release go run cmd/api/main.go`

By default, you can't pause subscription while in trial period. If you want to disable it
set the environment variable `ALLOW_PAUSE_ON_TRIAL` locally or in `docker-compose` file before run.

### Docker

Since multiple services are describe on docker-file, especify then when running:
```bash
docker-compose up -d --build database membershipapi
```

You can also use `make run`.

The API will keep restarting until it connects to the database.

## Documentation

You can get the API documentation as swagger by two means:

1. inside the `cmd/api/swagger` there will be a `swagger.jon` file.
2. after running the API, access `localhost:8081/swagger/swagger/` on a browser.

Also, there are a postman collection and environment that you can import. It's in `collection` folder.

## Tests

These tests relies heavly on functional tests, so the tests execution may take around 1 min depending on your setup.

In order to run tests, a database instance must be running.

If are in a environment with Make installed, run `make updbtest`.
Then, after the database is ready, run `make test`.

You want to do it manually, use the commands:
- `docker-compose up -d databasetest`
- `DB_HOST=localhost DB_PORT=5433 DB_NAME=membership DB_USER=tester DB_PW=secretpw go test -v -cover -count=1 ./...`

## Improvement

Other componets could be added in general. Eg.: middlware for authentication, propagation of logging and context, etc.

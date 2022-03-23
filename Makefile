.PHONY: runDbTest test testLocal testWithDb

runDbTest:
	@echo "setting up database with no permanent data for tests"
	docker-compose up -d databasetest

test:
	DB_HOST=database DB_PORT=5433 DB_NAME=membership DB_USER=postgres DB_PW=secretpwgo test -v -cover ./...

testLocal:
	test -v -cover ./...

testWithDb:
	runDbTest
	sleep 5
	test
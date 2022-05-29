dbContainer = timescale/timescaledb:latest-pg14
testDbPort = 5433
backendTest = go test ./... -v -count=1
frontendTest = cd web && npm i && npm test -- --watchAll=false
stopDevDb = ($(MAKE) stop-dev-db)

.PHONY: test
test: start-dev-db wait-db test-frontend test-backend-with-db-stop
	@echo All tests pass!

.PHONY: test-backend-with-db-stop
test-backend-with-db-stop:
	$(backendTest) || $(stopDevDb)

.PHONY: test-frontend-with-db-stop
test-frontend-with-db-stop: start-dev-db wait-db
	$(frontendTest) || (cd ../ && $(stopDevDb))

.PHONY: test-backend
test-backend: start-dev-db wait-db
	$(backendTest)  || $(stopDevDb)

.PHONY: test-frontend
test-frontend:
	$(frontendTest)

.PHONY: start-dev-db
start-dev-db:
	docker run -d --rm --name testdb -p $(testDbPort):5432 -e POSTGRES_PASSWORD=password $(dbContainer) \
	|| $(stopDevDb)

.PHONY: stop-dev-db
stop-dev-db:
	docker stop testdb

.PHONY: wait-db
wait-db:
	until docker run \
	--rm \
	--link testdb:pg \
	$(dbContainer) pg_isready \
		-U postgres \
		-h pg; do sleep 1; done

.PHONY: cover
cover:
	go test ./... -coverprofile cover.out && go tool cover -html=cover.out

.PHONY: build-docker
build-docker:
	docker build . -t $(TAG)

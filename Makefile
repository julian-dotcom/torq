dbContainer = timescale/timescaledb:latest-pg14
testDbPort = 5433

.PHONY: test
test: start-dev-db wait-db test-backend test-frontend stop-dev-db

.PHONY: test-backend
test-backend:
	go test ./... -v -count=1 || ($(MAKE) stop-dev-db && false)

.PHONY: start-dev-db
start-dev-db:
	docker run -d --rm --name testdb -p $(testDbPort):5432 -e POSTGRES_PASSWORD=password $(dbContainer) \
	|| $(MAKE) stop-dev-db

.PHONY: stop-dev-db
stop-dev-db:
	docker stop testdb

.PHONY: test-frontend
test-frontend:
	cd web && npm i && npm test -- --watchAll=false || (cd ../ && $(MAKE) stop-dev-db && false)

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

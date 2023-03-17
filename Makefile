dbContainer = timescale/timescaledb:latest-pg14
testDbPort = 5433
stopDevDb = ($(MAKE) stop-dev-db && false)
#Virtual Network - Frequency(every x seconds; default 1) of creating and paying invoices
virtual_network_invoice_freq = 30
#Virtual Network - Frequency(every x seconds; default 30) of sending coins to random address
virtual_network_send_coins_freq = 30
#Virtual Network - Frequency(every x minutes; default 10) of opening and closing random channels
virtual_network_open_close_chan_freq = 10
buildFrontend = cd web && npm install --legacy-peer-deps && npm run build && echo "Frontend Build Done"
frontendTest = cd web && npm test -- --watchAll=false
lintFrontend = cd web && npm run lint
lintBackend = go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && ~/go/bin/golangci-lint run
backendTest = go test ./... -v -count=1

.PHONY: test
test: lint start-dev-db wait-db test-backend-with-db-stop test-frontend-with-db-stop stop-dev-db
	@echo All tests pass!

.PHONY: test-backend-with-db-stop
test-backend-with-db-stop:
	$(backendTest) || $(stopDevDb)

.PHONY: test-frontend-with-db-stop
test-frontend-with-db-stop: buildFrontend
	$(frontendTest) || (cd ../ && $(stopDevDb))

.PHONY: test-backend
test-backend:
	$(backendTest)

.PHONY: test-frontend
test-frontend: buildFrontend
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

.PHONY: test-e2e
test-e2e:
	REACT_APP_E2E_TEST=true E2E=true go test -timeout 20m -v -count=1 ./test/e2e/lnd

.PHONY: test-e2e-debug
test-e2e-debug:
	REACT_APP_E2E_TEST=true E2E=true DEBUG=true go test -timeout 20m -v -count=1 ./test/e2e/lnd

.PHONY: create-dev-env
create-dev-env:
	go build ./virtual_network/torq_vn && go run ./virtual_network/torq_vn create --db true

.PHONY: start-dev-env
start-dev-env:
	go build ./virtual_network/torq_vn &&  go run ./virtual_network/torq_vn start --db true

.PHONY: stop-dev-env
stop-dev-env:
	go build ./virtual_network/torq_vn &&  go run ./virtual_network/torq_vn stop --db true

.PHONY: purge-dev-env
purge-dev-env:
	go build ./virtual_network/torq_vn &&  go run ./virtual_network/torq_vn purge --db true

# Start flow
.PHONY: start-dev-flow
start-dev-flow:
	DEBUG=true go run ./virtual_network/torq_vn flow --virtual_network_invoice_freq $(virtual_network_invoice_freq) --virtual_network_send_coins_freq $(virtual_network_send_coins_freq) --virtual_network_open_close_chan_freq $(virtual_network_open_close_chan_freq)

.PHONY: lint-backend
lint-backend: buildFrontend
	$(lintBackend)

.PHONY: lint-frontend
lint-frontend: buildFrontend
	$(lintFrontend)

.PHONY: lint
lint: lint-backend lint-frontend
	@echo Linting complete

.PHONY: generate-ts
generate-ts:
	go run cmd/torq/internal/generators/gen.go

.PHONY: buildFrontend
buildFrontend:
	$(buildFrontend)

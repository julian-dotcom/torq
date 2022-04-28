.PHONY: test
test: backend-test frontend-test

.PHONY: backend-test
backend-test:
	go test ./... -v

.PHONY: frontend-test
frontend-test:
	cd web && npm i && npm test -- --watchAll=false

.PHONY: cover
cover:
	go test ./... -coverprofile cover.out && go tool cover -html=cover.out

.PHONY: build-docker
build-docker:
	docker build . -t $(TAG)

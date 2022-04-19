
test:
	go test ./... -v

cover:
	go test ./... -coverprofile cover.out && go tool cover -html=cover.out

devcert:
	go run $(GOROOT)/src/crypto/tls/generate_cert.go --host localhost
	@echo "\n----\nRemember to allow the use of the unsigned certificate (from the organization Acme Co) in the browser."
	@echo "\nYou can manually visit localhost:50051 and change the trust settings\n---\n"

.PHONY: build-docker
build-docker:
	docker build . -t $(TAG)

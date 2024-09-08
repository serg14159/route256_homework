.PHONY: run
run:
	go run cmd/main/main.go

.PHONY: build
build:
	go build -o cart cmd/main/main.go

.PHONY: docker_build
docker_build:
	docker build -t cart:latest .

.PHONY: docker_run
docker_run:
	docker run -p 8082:8082 cart:latest

.PHONY: test_coverage
test_coverage: test coverage

.PHONY: test
test:
	go test $(shell go list ./... | grep -v /mock) -coverprofile=coverage.out

.PHONY: coverage
coverage:
	go tool cover -func=coverage.out
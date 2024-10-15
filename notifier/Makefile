# Run
.PHONY: run
run:
	go run cmd/main/main.go

# Tidy
.PHONY: tidy
tidy:
	go mod tidy

# Build
.PHONY: build
build:
	go build -o notifier cmd/main/main.go

# Docker
.PHONY: docker_build
docker_build:
	docker build -t notifier:latest .

.PHONY: docker_run
docker_run:
	docker run -p 8084:8084 notifier:latest

# Tests
.PHONY: test_all
test_all: cyclomatic_complexity cognitive_complexity test coverage

.PHONY: test
test:
	go test $(shell go list ./... | grep -vE '/(mock|vendor-proto)') -coverprofile=coverage.out

.PHONY: test_race
test_race:
	go test -race $(shell go list ./... | grep -vE '/(mock|vendor-proto)') -coverprofile=coverage.out

.PHONY: coverage
coverage:
	go tool cover -func=coverage.out

.PHONY: test_bench
test_bench:
	go test -bench=. ./...
	
# Lint
.PHONY: cognitive_complexity
cognitive_complexity:
	gocognit -over 10 -ignore "_test.go|Godeps|mock/|pkg/api/loms|vendor-proto/" .

.PHONY:	cyclomatic_complexity
cyclomatic_complexity:
	gocyclo -over 10 --ignore "_test.go|Godeps|mock/|pkg/api/loms|vendor-proto/" .
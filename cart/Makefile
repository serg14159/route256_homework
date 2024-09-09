# run
.PHONY: run
run:
	go run cmd/main/main.go

# build
.PHONY: build
build:
	go build -o cart cmd/main/main.go

# docker
.PHONY: docker_build
docker_build:
	docker build -t cart:latest .

.PHONY: docker_run
docker_run:
	docker run -p 8082:8082 cart:latest

# tests
.PHONY: test_all
test_all: cyclomatic_complexity cognitive_complexity test coverage

.PHONY: test
test:
	go test $(shell go list ./... | grep -v /mock) -coverprofile=coverage.out

.PHONY: coverage
coverage:
	go tool cover -func=coverage.out

.PHONY: test_bench
test_bench:
	go test -bench=. ./...
	
# lint
.PHONY: cognitive_complexity
cognitive_complexity:
	gocognit -over 10 -ignore "_test.go|Godeps|mock/" .

.PHONY:	cyclomatic_complexity
cyclomatic_complexity:
	gocyclo -over 10 --ignore "_test.go|Godeps|mock/" .
# Use bin in the current directory for installing plugins protoc
LOCAL_BIN:=$(CURDIR)/bin

# Add bin in the current directory to the PATH when starting the protocol
PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc

# Delete vendor-proto
.PHONY: .vendor-rm
.vendor-rm:
	rm -rf vendor-proto

# Installing proto descriptions of google/protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
	https://github.com/protocolbuffers/protobuf vendor-proto/protobuf &&\
	cd vendor-proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/protobuf/src/google/protobuf vendor-proto/google
	rm -rf vendor-proto/protobuf


# Installing proto descriptions to validate
vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
	https://github.com/bufbuild/protoc-gen-validate vendor-proto/tmp && \
	cd vendor-proto/tmp && \
	git sparse-checkout set --no-cone validate &&\
	git checkout
	mkdir -p vendor-proto/validate
	mv vendor-proto/tmp/validate vendor-proto/
	rm -rf vendor-proto/tmp


# Installing proto descriptions of google/googleapis
vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 	https://github.com/googleapis/googleapis vendor-proto/googleapis && \
 	cd vendor-proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor-proto/google
	mv vendor-proto/googleapis/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis

# Installing proto-protocol-gen-openapiv2/options
vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 	https://github.com/grpc-ecosystem/grpc-gateway vendor-proto/grpc-ecosystem && \
 	cd vendor-proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor-proto/protoc-gen-openapiv2
	mv vendor-proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor-proto/protoc-gen-openapiv2
	rm -rf vendor-proto/grpc-ecosystem

# Vendor of external proto files
vendor-proto: .vendor-rm  vendor-proto/google/protobuf vendor-proto/validate vendor-proto/google/api vendor-proto/protoc-gen-openapiv2/options

# Installing binary dependencies
.PHONY: bin-deps
bin-deps:
	$(info Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
	GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@v1.21.0 && \
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4 && \
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.19.1 && \
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.19.1 && \
	GOBIN=$(LOCAL_BIN) go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.5

# Generate code
.PHONY: protoc-generate
protoc-generate: bin-deps vendor-proto goprotos gwprotos validate openapiv2 tidy

# Common variables
LOMS_PROTO_PATH:="api/loms/v1"
LOMS_PKG_OUT_DIR=pkg/${LOMS_PROTO_PATH}
PROTO_FILE=api/loms/v1/loms.proto

.PHONY: goprotos
goprotos:	
	echo "Generating GO bindings"
	mkdir -p $(LOMS_PKG_OUT_DIR)
	protoc -I ${LOMS_PROTO_PATH} \
	       -I vendor-proto \
	       --plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
	       --go_out $(LOMS_PKG_OUT_DIR) \
	       --go_opt paths=source_relative \
	       --plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
	       --go-grpc_out $(LOMS_PKG_OUT_DIR) \
	       --go-grpc_opt paths=source_relative \
	       $(PROTO_FILE)

.PHONY: gwprotos
gwprotos:	
	echo "Generating gRPC Gateway bindings"
	mkdir -p $(LOMS_PKG_OUT_DIR)
	protoc -I ${LOMS_PROTO_PATH} \
	       -I vendor-proto \
	       --plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway \
	       --grpc-gateway_out $(LOMS_PKG_OUT_DIR) \
	       --grpc-gateway_opt logtostderr=true \
	       --grpc-gateway_opt paths=source_relative \
	       --grpc-gateway_opt generate_unbound_methods=true \
	       $(PROTO_FILE)

.PHONY: validate
validate:	
	echo "Generating validation bindings"
	mkdir -p $(LOMS_PKG_OUT_DIR)
	protoc -I ${LOMS_PROTO_PATH} \
	       -I vendor-proto \
	       --plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
	       --validate_out="lang=go,paths=source_relative:$(LOMS_PKG_OUT_DIR)" \
	       $(PROTO_FILE)

.PHONY: openapiv2
openapiv2:	
	echo "Generating OpenAPI v2 specs"
	mkdir -p api/openapiv2
	protoc -I ${LOMS_PROTO_PATH} \
	       -I vendor-proto \
	       --plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 \
	       --openapiv2_out api/openapiv2 \
	       --openapiv2_opt logtostderr=true \
	       $(PROTO_FILE)

# Tidy
.PHONY: tidy
tidy:
	go mod tidy

# Serve swagger
.PHONY: serve-swagger
serve-swagger:
	bin/swagger serve api/openapiv2/loms.swagger.json

# Run
.PHONY: run
run:
	go run cmd/main/main.go

# Build
.PHONY: build
build:
	go build -o loms cmd/main/main.go

# Docker
.PHONY: docker_build
docker_build:
	docker build -t loms:latest .

.PHONY: docker_run
docker_run:
	docker run -p 8082:8082 loms:latest

# Tests
.PHONY: test_all
test_all: cyclomatic_complexity cognitive_complexity test coverage

.PHONY: test
test:
	go test $(shell go list ./... | grep -vE '/(mock|pkg/api/loms|vendor-proto)') -coverprofile=coverage.out

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


# Migrations
.PHONY: install-goose
install-goose:
	$(info Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.22.1

.PHONY: create_migration_create_statuses_table
create_migration_create_statuses_table:
	bin/goose -dir migrations create create_statuses_table sql

.PHONY: create_migration_insert_initial_statuses_data
create_migration_insert_initial_statuses_data:
	bin/goose -dir migrations create insert_initial_statuses_data sql

.PHONY: create_migration_create_orders_table
create_migration_create_orders_table:
	bin/goose -dir migrations create create_orders_table sql

.PHONY: create_migration_create_items_table
create_migration_create_items_table:
	bin/goose -dir migrations create create_items_table sql

.PHONY: create_migration_create_stocks_table
create_migration_create_stocks_table:
	bin/goose -dir migrations create create_stocks_table sql

.PHONY: create_migration_insert_initial_stocks_data
create_migration_insert_initial_stocks_data:
	bin/goose -dir migrations create insert_initial_stocks_data sql

.PHONY: migrations_up
migrations_up:
	bin/goose -dir migrations postgres "postgres://user:password@localhost:5434/postgres?sslmode=disable" up

.PHONY: migrations_down
migrations_down:
	bin/goose -dir migrations postgres "postgres://user:password@localhost:5434/postgres?sslmode=disable" down

.PHONY: migrate
migrate:
	go run cmd/migration/main.go


.PHONY: sqlc_generate
sqlc_generate:
	sqlc generate
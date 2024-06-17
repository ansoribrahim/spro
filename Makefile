.PHONY: clean all init generate generate_mocks

# Define directories where mocks will be generated
REPOSITORY_DIRS := repository
SERVICE_DIRS := service

# Define targets for mocks generation in repository and service folders
REPOSITORY_INTERFACES_GO_FILES := $(shell find $(REPOSITORY_DIRS) -name "interfaces.go")
REPOSITORY_INTERFACES_GEN_GO_FILES := $(REPOSITORY_INTERFACES_GO_FILES:%.go=%.mock.gen.go)

SERVICE_INTERFACES_GO_FILES := $(shell find $(SERVICE_DIRS) -name "interfaces.go")
SERVICE_INTERFACES_GEN_GO_FILES := $(SERVICE_INTERFACES_GO_FILES:%.go=%.mock.gen.go)

all: build/main

build/main: cmd/main.go generated
	@echo "Building..."
	go build -o $@ $<

cover:
	go tool cover -html=coverage.out

clean:
	rm -rf generated

init: clean generate
	go mod tidy
	go mod vendor

test:
	go clean -testcache
	go test -short -coverprofile coverage.out -v ./...
	@# Remove lines from the coverage report that match the pattern *.gen.*
	@grep -v '\.gen\.' coverage.out > coverage.filtered.out
	@mv coverage.filtered.out coverage.out
	@echo "===================="
	@echo "Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"
	@echo "===================="

test_api:
	go clean -testcache
	go test ./tests/...

generate: generated generate_mocks

generated: api.yml
	@echo "Generating files..."
	mkdir generated || true
	oapi-codegen --package generated -generate types,server,spec $< > generated/api.gen.go
	@echo "Adding validation tags..."
	go run config/add_validation_tags.go

generate_mocks: $(REPOSITORY_INTERFACES_GEN_GO_FILES) $(SERVICE_INTERFACES_GEN_GO_FILES)

$(REPOSITORY_INTERFACES_GEN_GO_FILES): %.mock.gen.go: %.go
	@echo "Generating mocks $@ for $<"
	mockgen -source=$< -destination=$@ -package=$(shell basename $(dir $<))

$(SERVICE_INTERFACES_GEN_GO_FILES): %.mock.gen.go: %.go
	@echo "Generating mocks $@ for $<"
	mockgen -source=$< -destination=$@ -package=$(shell basename $(dir $<))

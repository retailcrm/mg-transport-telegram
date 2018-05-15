ROOT_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRC_DIR=$(ROOT_DIR)
CONFIG_FILE=$(ROOT_DIR)/config.yml
CONFIG_TEST_FILE=$(ROOT_DIR)/config_test.yml
BIN=$(ROOT_DIR)/mg-telegram
REVISION=$(shell git describe --tags 2>/dev/null || git log --format="v0.0-%h" -n 1 || echo "v0.0-unknown")

ifndef GOPATH
    $(error GOPATH must be defined)
endif

export GOPATH := $(GOPATH):$(ROOT_DIR)

fmt:
	@echo "==> Running gofmt"
	@gofmt -l -s -w $(SRC_DIR)

install: fmt
	@echo "==> Running go get"
	$(eval DEPS:=$(shell cd $(SRC_DIR) \
	 	&& go list -f '{{join .Imports "\n"}}{{ "\n" }}{{join .TestImports "\n"}}' ./... \
		| sort | uniq | tr '\r' '\n' | paste -sd ' ' -))
	@go get -d -v $(DEPS)

build: install
	@echo "==> Building"
	@go build -o $(BIN) -ldflags "-X common.build=${REVISION}" .
	@echo $(BIN)

migrate: build
	@${BIN} --config $(CONFIG_FILE) migrate -p ./migrations/

run: migrate
	@echo "==> Running"
	@${BIN} --config $(CONFIG_FILE) run
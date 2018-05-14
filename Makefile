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

build: install
	@echo "==> Building"
	@go build -o $(BIN) -ldflags "-X common.build=${REVISION}" .
	@echo $(BIN)

run: build
	@echo "==> Running"
	@${BIN}

install: fmt
	@echo "==> Running go get"
	@go get -v -u gopkg.in/yaml.v2 github.com/op/go-logging github.com/jinzhu/gorm \
	github.com/jinzhu/gorm/dialects/postgres github.com/retailcrm/api-client-go/v5 \
	github.com/golang-migrate/migrate github.com/golang-migrate/migrate/database/postgres \
	github.com/jessevdk/go-flags github.com/go-telegram-bot-api/telegram-bot-api
	@gofmt -l -s -w $(SRC_DIR)


fmt:
	@echo "==> Running gofmt"
	@gofmt -l -s -w $(SRC_DIR)

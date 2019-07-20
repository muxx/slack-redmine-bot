ROOT_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CONFIG_FILE=$(ROOT_DIR)/config.yml
BIN=$(ROOT_DIR)/bin/slack-redmine-bot
REVISION=$(shell git describe --tags 2>/dev/null || git log --format="v0.0-%h" -n 1 || echo "v0.0-unknown")

run:
	@echo "==> Running"
	@${BIN} --config $(CONFIG_FILE)

build: deps fmt build_only

build_only:
	@echo "==> Building"
	@cd $(ROOT_DIR) && CGO_ENABLED=0 go build -o $(BIN) -ldflags "-X common.build=${REVISION}" .
	@echo $(BIN)

fmt:
	@echo "==> Running gofmt"
	@gofmt -l -s -w $(ROOT_DIR)

deps:
	@echo "==> Installing dependencies"
	@go mod tidy

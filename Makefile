MQTT_ACTIONS_BIN := bin/mqtt-actions

GO ?= go
GO_MD2MAN ?= go-md2man

VERSION := $(shell cat VERSION)
USE_VENDOR =
LOCAL_LDFLAGS = -buildmode=pie -ldflags "-X=github.com/thkukuk/mqtt-actions/pkg/mqtt-actions.Version=$(VERSION)"

.PHONY: all api build vendor
all: dep build

dep: ## Get the dependencies
	@$(GO) get -v ./...

update: ## Get and update the dependencies
	@$(GO) get -v -u ./...

tidy: ## Clean up dependencies
	@$(GO) mod tidy

vendor: dep ## Create vendor directory
	@$(GO) mod vendor

build: ## Build the binary files
	$(GO) build -v -o $(MQTT_ACTIONS_BIN) $(USE_VENDOR) $(LOCAL_LDFLAGS) ./cmd/mqtt-actions

clean: ## Remove previous builds
	@rm -f $(MQTT_ACTIONS_BIN)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: release
release: ## create release package from git
	git clone https://github.com/thkukuk/mqtt-actions
	mv mqtt-actions mqtt-actions-$(VERSION)
	sed -i -e 's|USE_VENDOR =|USE_VENDOR = -mod vendor|g' mqtt-actions-$(VERSION)/Makefile
	make -C mqtt-actions-$(VERSION) vendor
	cp VERSION mqtt-actions-$(VERSION)
	tar --exclude .git -cJf mqtt-actions-$(VERSION).tar.xz mqtt-actions-$(VERSION)
	rm -rf mqtt-actions-$(VERSION)

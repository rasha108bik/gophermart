LOCAL_BIN=$(CURDIR)/bin

default: help

.PHONY: help
help: ## help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: app-run
app-run: ## run app
	go run $(PWD)/cmd/gophermart/main.go
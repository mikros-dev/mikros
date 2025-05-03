default: generate

generate: ## Generate code from protobuf files (default option)
	@.scripts/generate.sh

clean: ## Clean all generated artifacts
	@rm -rf gen

setup: ## Install required tools and dependencies
	@.scripts/setup.sh

go-check: ## Compile all golang generated modules
	@.scripts/go.sh -b

test-check: ## Compile all golang tests generated modules
	@.scripts/go.sh -t

go-get: ## Update golang dependencies
	@.scripts/go.sh -u

mock-check: ## Compile generated mocks
	@.scripts/go.sh -m

check: ## Execute all checks
	@$(MAKE) go-check
	@$(MAKE) test-check
	@$(MAKE) mock-check

help: ## Show all available options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: generate clean help setup go-check test-check go-get mock-check check

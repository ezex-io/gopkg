PACKAGES := logger util testsuite env middleware/http-mdl
ROOT_DIR := $(shell pwd)
LINT_CONFIG := $(ROOT_DIR)/.golangci.yml

check:
	@for pkg in $(PACKAGES); do \
		echo "Linting $$pkg package..."; \
		cd $(ROOT_DIR)/$$pkg && golangci-lint run --config $(LINT_CONFIG); \
	done

test:
	@for pkg in $(PACKAGES); do \
		echo "Testing $$pkg package..."; \
		cd $(ROOT_DIR)/$$pkg && go test ./...; \
	done

update:
	@for pkg in $(PACKAGES); do \
		echo "Updating $$pkg dependencies..."; \
		cd $(ROOT_DIR)/$$pkg && go get -u ./... ; \
		cd $(ROOT_DIR)/$$pkg && go mod tidy; \
	done

.PHONY: check test update

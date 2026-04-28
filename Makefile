.PHONY: test test-coverage coverage-html coverage-ci lint examples test-all

test:
	@echo "[*] Running unit tests..."
	@mkdir -p coverage
	go test ./... -coverpkg=./... -coverprofile=coverage/coverage.cov
	go tool cover -func=coverage/coverage.cov

test-coverage: test
	@echo ""
	@echo "[*] Coverage Summary:"
	@go tool cover -func=coverage/coverage.cov | tail -1

coverage-html: test
	@echo "[*] Generating HTML coverage report..."
	go tool cover -html=coverage/coverage.cov -o coverage/coverage.html
	@echo "[+] HTML report generated: coverage/coverage.html"

coverage-ci:
	@mkdir -p coverage
	go test ./... -coverpkg=./... -coverprofile=coverage/coverage.cov -covermode=atomic
	go tool cover -func=coverage/coverage.cov

lint:
	@echo "[*] Running linters..."
	golangci-lint run ./... 2>/dev/null || go vet ./...

# Build all examples
examples:
	@echo "Building examples..."
	@for f in examples/*/; do \
		echo "Building $$(basename $$f)..."; \
		go build $$f*.go || exit 1; \
	done
	@echo "All examples compile successfully!"

# Run unit tests + functional tests (all examples)
test-all: test examples
	@echo ""
	@echo "[*] Running functional tests on all examples..."
	@chmod +x scripts/test-all.sh
	@bash scripts/test-all.sh

# Run a specific example
run-%:
	@echo "Running example $*..."
	go run examples/$*/main.go

# Clean
clean:
	@rm -f examples/*/main
	@echo "Cleaned."


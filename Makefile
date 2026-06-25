ROLE := config
.PHONY: build test test-standalone-layout test-build-gate
build:
	go build -o bin/cofiswarm-config ./cmd/cofiswarm-config
test: test-standalone-layout test-build-gate
test-standalone-layout:
	./test/scripts/assert-layout.sh $(ROLE)
# Asserts the Go build/split are byte-identical to the committed swarm-config.json + config/agents.
test-build-gate:
	go test ./...

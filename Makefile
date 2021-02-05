# Run the entire test suite
.PHONY: test
test:
	find . -name go.mod -execdir go test ./... -cover -race -v \;

# Run the entire test suite
.PHONY: test
test:
	find . -name go.mod -execdir go test ./... -cover -race \;

.PHONY: bump
bump:
	sed -E "s/go.krak3n.io\/fido v[0-9]+[0-9]*(\.[0-9]+){2}/go.krak3n.io\/fido v$VERSION/g" <<< 'go.krak3n.io/fido v0.0.0'

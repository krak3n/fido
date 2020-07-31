COVEROUT  ?= coverage.out
COVERMODE ?= atomic

test:
	go test -race -coverprofile="${COVEROUT}" -covermode="${COVERMODE}" ./...

coverage: test
	go tool cover -html=${COVEROUT}

COVEROUT  ?= coverage.out
COVERMODE ?= atomic
GODOCPORT ?= 5000

test:
	go test -race -coverprofile="${COVEROUT}" -covermode="${COVERMODE}" ./...

coverage: test
	go tool cover -html=${COVEROUT}

godoc:
	godoc -http=:${GODOCPORT}

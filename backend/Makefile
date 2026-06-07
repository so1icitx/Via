.PHONY: test test-cover test-cover-html

test:
	go test ./...

test-cover:
	go test $$(go list ./internal/... | grep -v testutil) -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1

test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
.PHONY: fmt test

fmt:
	gofmt -s -w .

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

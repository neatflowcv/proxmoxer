.PHONY: build
build:
	go build ./cmd/...

.PHONY: fix
fix:
	# go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	golangci-lint run --fix

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor

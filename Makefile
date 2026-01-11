.PHONY: fix
fix:
	go tool -modfile=tools/go.mod golangci-lint run --fix

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor
	make -C tools update
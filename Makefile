.PHONY: test

GO=CGO_ENABLED=1 GO111MODULE=on go

test:
	$(GO) test ./... -cover -count=1
	$(GO) vet ./...
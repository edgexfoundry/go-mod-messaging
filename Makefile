.PHONY: test unittest lint

ARCH=$(shell uname -m)
GO=go

tidy:
	go mod tidy

TAGS=include_nats_messaging

unittest:
	$(GO) test -tags=$(TAGS) -race ./... -coverprofile=coverage.out ./...

lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run make install-lint"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi

install-lint:
	sudo curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.51.2

test: unittest lint
	$(GO) vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]

BENCHCNT := 5 # number of times to run each benchmark with `make bench`

bench:
	go test -tags=$(TAGS) -bench=. -count $(BENCHCNT) -run=^# ./...

vendor:
	go mod vendor

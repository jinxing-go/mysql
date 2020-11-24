GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")

.PHONY: test
test:
	$(GO) test -v -short -covermode=count -coverprofile=cover.out ./...
	$(GO) tool cover -func=cover.out -o cover_total.out
	$(GO) tool cover -html=cover.out -o cover.html

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)
	php gen
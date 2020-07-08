GO_BUILD_ARGS = -mod=readonly -v -ldflags="-s -w"

.PHONY: all
all: algolia checker packages autoupdate kv

.PHONY: algolia
algolia:
	go build $(GO_BUILD_ARGS) -o bin/algolia ./cmd/algolia

.PHONY: checker
checker:
	go build $(GO_BUILD_ARGS) -o bin/checker ./cmd/checker

.PHONY: packages
packages:
	go build $(GO_BUILD_ARGS) -o bin/packages ./cmd/packages

.PHONY: autoupdate
autoupdate:
	go build $(GO_BUILD_ARGS) -o bin/autoupdate ./cmd/autoupdate

.PHONY: kv
autoupdate:
	go build $(GO_BUILD_ARGS) -o bin/kv ./cmd/kv

.PHONY: clean
clean:
	rm -rfv bin/*

.PHONY: test
test: clean checker
	go test -v ./test/...

.PHONY: lint
lint:
	go get -u golang.org/x/lint/golint
	$(GOPATH)/bin/golint ./...

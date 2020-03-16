GO_BUILD_ARGS = -mod=readonly -v -ldflags="-s -w"
all: algolia checker packages autoupdate

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

all: algolia checker packages autoupdate

.PHONY: algolia
algolia:
	go build -v -ldflags="-s -w" -o bin/algolia ./cmd/algolia

.PHONY: checker
checker:
	go build -v -ldflags="-s -w" -o bin/checker ./cmd/checker

.PHONY: packages
packages:
	go build -v -ldflags="-s -w" -o bin/packages ./cmd/packages

.PHONY: autoupdate
autoupdate:
	go build -v -ldflags="-s -w" -o bin/autoupdate ./cmd/autoupdate

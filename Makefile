GO_BUILD_ARGS = -mod=readonly -v -ldflags="-s -w"

.PHONY: all
all: bin/process-version-host bin/git-sync \
	functions/check-pkg-updates/check-pkg-updates.zip \
	functions/process-version/process-version.zip functions/kv-pump/kv-pump.zip

bin/git-sync:
	go build $(GO_BUILD_ARGS) -o bin/git-sync ./cmd/git-sync

bin/process-version-host:
	go build $(GO_BUILD_ARGS) -o bin/process-version-host ./cmd/process-version-host

.PHONY: schema
schema:
	./bin/packages human > schema_human.json
	./bin/packages non-human > schema_non_human.json

.PHONY: clean
clean:
	rm -rfv bin/*
	rm -rfv functions/*/*.zip

.PHONY: test
test: clean checker
	go test -v ./test/...

.PHONY: lint
lint:
	go get -u golang.org/x/lint/golint
	$(GOPATH)/bin/golint ./...

.PHONY: dev
dev: autoupdate
	docker build -t cdnjs-dev -f ./dev/Dockerfile .
	docker run -it cdnjs-dev

functions/process-version/process-version.zip:
	make -C ./functions/process-version process-version.zip

functions/check-pkg-updates/check-pkg-updates.zip:
	make -C ./functions/check-pkg-updates check-pkg-updates.zip

functions/kv-pump/kv-pump.zip:
	make -C ./functions/kv-pump kv-pump.zip

BINARY  = pager
DIST    = dist
VERSION ?= v0.1.0

.PHONY: build build-all release run-override run-oncall test test-v vet tidy clean check

## Build the binary (current platform)
build:
	go build -o $(BINARY) .

## Cross-compile binaries for all platforms into dist/
build-all:
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(BINARY)-windows-amd64.exe .

## Create a GitHub release and upload binaries (requires gh CLI)
## Usage: make release VERSION=v0.2.0
release: build-all
	git tag $(VERSION)
	git push origin $(VERSION)
	gh release create $(VERSION) \
		--title "$(VERSION) — Initial Release" \
		--notes-file RELEASE_NOTES.md \
		$(DIST)/$(BINARY)-darwin-arm64 \
		$(DIST)/$(BINARY)-darwin-amd64 \
		$(DIST)/$(BINARY)-linux-amd64 \
		$(DIST)/$(BINARY)-linux-arm64 \
		$(DIST)/$(BINARY)-windows-amd64.exe

## Run the override command
run-override: build
	./$(BINARY) override

## Run the oncall command
run-oncall: build
	./$(BINARY) oncall

## Run all tests
test:
	go test ./...

## Run all tests with verbose output
test-v:
	go test -v ./...

## Run go vet
vet:
	go vet ./...

## Tidy module dependencies
tidy:
	go mod tidy

## Remove built binary and dist/
clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

## Build + vet
check: vet build

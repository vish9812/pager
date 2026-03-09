BINARY = pager

.PHONY: build run-override run-oncall test vet tidy clean

## Build the binary
build:
	go build -o $(BINARY) .

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

## Remove built binary
clean:
	rm -f $(BINARY)

## Build + vet
check: vet build

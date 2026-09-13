# libtheme-css. `make` runs the checks; `make build` puts the console tool
# in bin/, which is gitignored.
.PHONY: all fmt vet test build
all: fmt vet test
build:
	go build -o bin/libtheme ./cmd/libtheme
fmt:
	gofmt -l -w .
vet:
	go vet ./...
test:
	go test ./...

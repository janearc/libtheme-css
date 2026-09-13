# libtheme-css. `make` runs the checks; `make build` puts every command
# in bin/, which is gitignored; `make clean` takes bin/ away again.
.PHONY: all fmt vet test build clean
all: fmt vet test
fmt:
	gofmt -l -w .
vet:
	go vet ./...
test:
	go test ./...
build:
	mkdir -p bin
	go build -o bin/ ./cmd/...
clean:
	rm -rf bin

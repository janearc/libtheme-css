# libtheme-css. `make` runs the checks; `make build` puts every command
# in bin/, which is gitignored; `make clean` takes bin/ away again.
.PHONY: all fmt vet test build clean visualtest
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
# everything the library can derive without being told a colour, painted,
# so a change lower down is a change you can see.
visualtest: build
	./bin/libtheme known

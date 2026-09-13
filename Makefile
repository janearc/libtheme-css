# libtheme-css. no binaries yet; this is a library. `make` runs the checks.
.PHONY: all fmt vet test
all: fmt vet test
fmt:
	gofmt -l -w .
vet:
	go vet ./...
test:
	go test ./...

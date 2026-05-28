.PHONY: build test test-verbose check test-integration vet fmt coverage run clean deb

build:
	go build -o beep-analytics ./cmd/beep-analytics

test:
	go test ./...

test-verbose:
	go test ./... -v

check: fmt vet test test-integration

test-integration:
	go test -tags=integration -v ./tests/

vet:
	go vet ./...

fmt:
	go fmt ./...

coverage:
	go test -coverprofile=coverage/coverage.out ./... && go tool cover -html=coverage/coverage.out -o coverage/index.html

run:
	go run ./cmd/beep-analytics serve

clean:
	rm -f beep-analytics coverage/coverage.out coverage/index.html

deb:
	./scripts/build-deb.sh

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# OAuth credentials - set via environment or .env file
YOY_CLIENT_ID     ?= $(shell echo $$YOY_CLIENT_ID)
YOY_CLIENT_SECRET ?= $(shell echo $$YOY_CLIENT_SECRET)

LDFLAGS := -X github.com/Softorize/yoy/internal/version.Version=$(VERSION) \
           -X github.com/Softorize/yoy/internal/version.Commit=$(COMMIT) \
           -X github.com/Softorize/yoy/internal/version.BuildDate=$(DATE)

ifneq ($(YOY_CLIENT_ID),)
LDFLAGS += -X github.com/Softorize/yoy/internal/auth.oauthClientID=$(YOY_CLIENT_ID)
endif
ifneq ($(YOY_CLIENT_SECRET),)
LDFLAGS += -X github.com/Softorize/yoy/internal/auth.oauthClientSecret=$(YOY_CLIENT_SECRET)
endif

.PHONY: build test lint clean install fmt vet

build:
	go build -ldflags "$(LDFLAGS)" -o bin/yoy .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./... -v -count=1

test-short:
	go test ./... -short -count=1

lint: vet
	@which staticcheck > /dev/null 2>&1 || echo "Install staticcheck: go install honnef.co/go/tools/cmd/staticcheck@latest"
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || true

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin/ dist/

tidy:
	go mod tidy

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

all: fmt vet test build

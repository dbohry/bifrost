.PHONY: help web build run dev test lint size e2e image

GO_PKGS := ./cmd/... ./internal/...
VERSION ?= dev

help:
	@echo "make web     build the frontend into internal/webui/dist"
	@echo "make build   build the frontend, then bin/bifrost"
	@echo "make run     run bin/bifrost with ./data as the data directory"
	@echo "make dev     run the Go server (:8080) and the Vite dev server (:5173)"
	@echo "make test    run the Go tests"
	@echo "make lint    gofmt, go vet and svelte-check"
	@echo "make size    check the frontend bundle against its size budget"
	@echo "make e2e     browser checks (run make build first)"
	@echo "make image   build the Docker image as bifrost:dev"

web:
	cd frontend && npm ci && npm run build

build: web
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/bifrost ./cmd/bifrost

run: build
	BIFROST_DATA_DIR=./data BIFROST_ADDR=:8080 ./bin/bifrost

dev:
	cd frontend && npm install
	@trap 'kill 0' INT TERM EXIT; \
	BIFROST_DATA_DIR=./data go run ./cmd/bifrost & \
	cd frontend && npm run dev

test:
	go test -race $(GO_PKGS)

lint:
	@test -z "$$(gofmt -l cmd internal)" || { gofmt -l cmd internal; echo "gofmt: files above need formatting"; exit 1; }
	go vet $(GO_PKGS)
	cd frontend && npm run check

size:
	cd frontend && npm run size

e2e:
	e2e/run.sh

image:
	docker build --build-arg VERSION=$(VERSION) -t bifrost:dev .

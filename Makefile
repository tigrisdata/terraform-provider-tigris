GOFMT_FILES?=$$(find . -name '*.go')
PKG_NAME=tigris
VERSION?=$(shell git describe --tags --always)
INCLUDE_VERSION_IN_FILENAME?=false

default: build

install: vet fmtcheck
	go install -ldflags="-X github.com/tigrisdata/terraform-provider-tigris/main.version=$(VERSION)"

build: vet
	@if $(INCLUDE_VERSION_IN_FILENAME); then \
	    go build -ldflags="-X github.com/tigrisdata/terraform-provider-tigris/main.version=$(VERSION)" -o terraform-provider-tigris_$(VERSION); \
		echo "==> Successfully built terraform-provider-tigris_$(VERSION)"; \
	else \
		go build -ldflags="-X github.com/tigrisdata/terraform-provider-tigris/main.version=$(VERSION)" -o terraform-provider-tigris; \
		echo "==> Successfully built terraform-provider-tigris"; \
	fi

lint: tools terraform-provider-lint golangci-lint

terraform-provider-lint: tools
	$$(go env GOPATH)/bin/tfproviderlintx \
	 -R001=false \
	 -R003=false \
	 -R012=false \
	 -R018=false \
	 -S006=false \
	 -S014=false \
	 -S020=false \
	 -S022=false \
	 -S023=false \
	 -AT001=false \
	 -AT002=false \
	 -AT003=false \
	 -AT006=false \
	 -AT012=false \
	 -R013=false \
	 -XAT001=false \
	 -XR001=false \
	 -XR003=false \
	 -XR004=false \
	 -XS001=false \
	 -XS002=false \
	 ./...

vet:
	@echo "==> Running go vet ."
	@go vet ./... ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

golangci-lint:
	@golangci-lint run ./internal/... --config .golintci.yml

tools:
	@echo "==> Installing development tooling..."
	go generate -tags tools tools/tools.go

docs: tools
	@sh -c "'$(CURDIR)/scripts/generate-docs.sh'"

.PHONY: build install lint terraform-provider-lint vet fmt golangci-lint tools
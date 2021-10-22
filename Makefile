NAME=dbt-cloud
BINARY=terraform-provider-$(NAME)
VERSION=$(shell cat VERSION)

default: install

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go get golang.org/x/tools/cmd/goimports

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64
	mv $(BINARY) ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64/$(BINARY)

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

test: deps
	go test -mod=readonly ./...

check-docs: docs
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

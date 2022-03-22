NAME=dbt-cloud
VERSION=$(shell cat VERSION)
BINARY=terraform-provider-$(NAME)_v$(VERSION)

default: install

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go install golang.org/x/tools/cmd/goimports@latest

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.terraform.d/plugins
	mv ./$(BINARY) $(HOME)/.terraform.d/plugins/$(BINARY)

docs:
	go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

test: deps
	go test -mod=readonly -count=1 ./...

test-acceptance: deps
	TF_ACC=1 go test -mod=readonly -count=1 ./...

check-docs: docs
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

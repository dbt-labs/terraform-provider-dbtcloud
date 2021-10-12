NAME=dbt-cloud
BINARY=terraform-provider-$(NAME)
VERSION=$(shell cat VERSION)

default: install

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64
	mv $(BINARY) ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64/$(BINARY)

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

test: fmt deps
	go test

check-docs: docs
	git diff --exit-code -- docs

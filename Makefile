NAME=dbt-cloud
BINARY=terraform-provider-$(NAME)
VERSION=0.1

default: install

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64
	mv $(BINARY) ~/.terraform.d/plugins/gthesheep/dbt_cloud/0.1/darwin_amd64/$(BINARY)

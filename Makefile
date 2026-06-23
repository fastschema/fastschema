testpkg:
	./tests/test.sh ./pkg/...

test:
	./tests/test.sh

validate-example-schemas:
	@go run ./docs/schemas/cmd docs/schemas/
.PHONY: validate-example-schemas

lint:
	golangci-lint run

dev:
	mkdir -p ./data/tmp
	air -c ./.air.toml start .

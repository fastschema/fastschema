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

builddash: export DIST_DIR=../../dash
builddash:
	git submodule update --remote --merge
	cd pkg/dash && yarn install && yarn build

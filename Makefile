testpkg:
	./tests/test.sh ./pkg/...

test:
	./tests/test.sh

lint:
	golangci-lint run

dev:
	mkdir -p ./data/tmp
	air -c ./.air.toml start .

builddash: export DIST_DIR=../../dash
builddash:
	git submodule update --remote --merge
	cd pkg/dash && yarn install && yarn build

testpkg:
	./tests/test.sh ./pkg/...

test:
	./tests/test.sh

lint:
	golangci-lint run

dev:
	air -c ./.air.toml start .

builddash: export DIST_DIR=../../dash
builddash:
	git submodule update --init --recursive
	cd pkg/dash && yarn install && yarn build

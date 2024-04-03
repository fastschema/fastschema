testpkg:
	./tests/test.sh ./pkg/...

test:
	./tests/test.sh

lint:
	golangci-lint run

dev:
	air -c ./.air.toml start .

builddash:
	cd pkg/dash && yarn install && yarn build

.PHONY: verify
verify:
	golint ./cmd/... ./pkg/... && go vet ./cmd/... ./pkg/...

.PHONY: tests
tests:
	docker-compose build tests && docker-compose run tests

.PHONY: run
run:
	docker-compose build rediproxy && docker-compose up rediproxy

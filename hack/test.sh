#!/bin/sh

go test -v ./pkg/api/...
go test -v -bench=. ./pkg/cache/...

# Uses the redis-url flag to run tests
# against the redis service.
go test -v ./pkg/service/... $1

# e2e test scenario
go test -v ./tests/... $1 $2

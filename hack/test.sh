#!/bin/sh

go test -v ./pkg/api/...
go test -v ./pkg/cache/...

# Uses the redis-url flag to run tests
# against the redis service.
go test -v ./pkg/service/... $1

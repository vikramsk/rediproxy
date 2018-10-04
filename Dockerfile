FROM golang:1.10.4-alpine3.8 as builder

COPY . $GOPATH/src/github.com/vikramsk/rediproxy
WORKDIR $GOPATH/src/github.com/vikramsk/rediproxy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/rediproxy ./cmd/rediproxy 

FROM scratch as release
COPY --from=builder /go/bin/rediproxy /usr/bin/rediproxy
ENTRYPOINT ["/usr/bin/rediproxy"]

FROM golang:alpine as builder

COPY . $GOPATH/src/github.com/vikramsk/rediproxy

WORKDIR $GOPATH/src/github.com/vikramsk/rediproxy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/rediproxy ./cmd/rediproxy 
FROM scratch

COPY --from=builder /go/bin/rediproxy /usr/bin/rediproxy

ENTRYPOINT ["/usr/bin/rediproxy"]

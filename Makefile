finbin.amd64.linux:	cmd/finbin/*.go
	GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $@ ./cmd/finbin

finbin:	cmd/finbin/*.go
	GO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o $@ ./cmd/finbin

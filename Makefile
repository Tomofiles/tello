GOPATH=$(shell pwd)
GOBIN=$(shell pwd)/bin
GOFILES=tomofiles/tello
GONAME=tello

all: install run

install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

run:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go run $(GOFILES)

clean:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

deps:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get \
		github.com/gorilla/websocket

test:
	@GOPATH=${GOPATH} GOBIN=${GOBIN} go test -v tomofiles/...

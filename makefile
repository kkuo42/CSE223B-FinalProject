.PHONY: all fmt tags doc rall test testv lc 

all:
	go install ./...

rall:
	go build -a ./...

fmt:
	gofmt -s -w -l .

tags:
	gotags -R . > tags

test:
	go test ./...

testv:
	go test -v ./...

lc:
	wc -l `find . -name "*.go"`

doc:
	godoc -http=:8000

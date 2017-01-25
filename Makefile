test:
	echo "Not implemented... yet :("

lint:
	gometalinter --config .linter.conf

deps:
	glide install

dev-deps: deps
	go get github.com/alecthomas/gometalinter
	gometalinter --install

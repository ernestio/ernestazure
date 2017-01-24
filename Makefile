test:
	echo "Not implemented... yet :("

lint:
	gometalinter --config .linter.conf

deps:
	echo "Not implemented... yet :("

dev-deps:
	go get -d github.com/Azure/azure-sdk-for-go
	go get github.com/alecthomas/gometalinter
	gometalinter --install

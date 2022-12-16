build:
	go get
	go-bindata -o bin.go assets/
	go install github.com/go-bindata/go-bindata/...@latest
	go install
	go build
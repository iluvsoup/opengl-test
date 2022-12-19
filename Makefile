build:
	go get
	go install github.com/go-bindata/go-bindata/...@latest
	go-bindata -o bin.go assets/
	go install
	go build

run:
	go get
	go install github.com/go-bindata/go-bindata/...@latest
	go-bindata -o bin.go assets/
	go install
	go run main
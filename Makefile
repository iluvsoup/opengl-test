build:
	go install github.com/go-bindata/go-bindata/...@latest
	go-bindata -o src/bin.go -pkg main assets/
	go install main/src
	go build -o dist/build main/src

run:
	go install github.com/go-bindata/go-bindata/...@latest
	go-bindata -o src/bin.go assets/
	go install main/src
	go run main/src
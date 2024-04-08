run:
	go run .

test:
	go test -v ./...

build:
	go build -o bin/wsgateway .

build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/wsgateway-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/wsgateway-darwin-arm64 .
	GOOS=linux GOARCH=arm go build -o bin/wsgateway-linux-arm .
	GOOS=linux GOARCH=arm64 go build -o bin/wsgateway-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o bin/wsgateway-windows-amd64 .

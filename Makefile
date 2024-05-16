run:
	cd src && go run -race . -config=../config.yml

test:
	cd src && go test -v ./...

build:
	cd src && go build -o ../bin/wsgateway .

build-all:
	cd src && GOOS=darwin GOARCH=amd64 go build -o ../bin/wsgateway-darwin-amd64 .
	cd src && GOOS=darwin GOARCH=arm64 go build -o ../bin/wsgateway-darwin-arm64 .
	cd src && GOOS=linux GOARCH=arm go build -o ../bin/wsgateway-linux-arm .
	cd src && GOOS=linux GOARCH=arm64 go build -o ../bin/wsgateway-linux-arm64 .
	cd src && GOOS=linux GOARCH=amd64 go build -o ../bin/wsgateway-linux-amd64 .
	# cd src && GOOS=windows GOARCH=amd64 go build -o ../bin/wsgateway-windows-amd64.exe .

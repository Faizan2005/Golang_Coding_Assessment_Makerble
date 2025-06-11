build:	
	@go build -o bin/app

run: build
	@DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_NAME=hospital ./bin/app 	

test:
	@go test -v ./...

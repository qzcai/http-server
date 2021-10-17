tag=v1.0

build:
	echo "building http-server binary"
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64 .

release: build
	echo "building http-server container"
	docker build -t caiqingzhong/http-server:${tag} .

push: release
	echo "pushing caiqingzhong/http-server"
	docker push caiqingzhong/http-server:${tag}

clean:
	go clean
	rm bin/amd64/http-server
	rmdir -p bin/amd64
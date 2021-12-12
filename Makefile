tag=v2.0

build:
	echo "building http-server container"
	docker build -t qzcai/http-server:${tag} .

push: build
	echo "pushing caiqingzhong/http-server"
	docker push qzcai/http-server:${tag}
FROM ubuntu
ENV VERSION=1.0
COPY bin/amd64/http-server /http-server
ENTRYPOINT /http-server
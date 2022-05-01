FROM golang:1.17-alpine as builder

RUN apk add --no-cache --update alpine-sdk=1.0-r1 \
    git=2.36.0-r0 \
    make=4.3-r0 \
    bash=5.1.16-r2 \
    protobuf-dev=3.18.1-r2 \
    protoc=3.18.1-r2

CMD ["/bin/sh"]

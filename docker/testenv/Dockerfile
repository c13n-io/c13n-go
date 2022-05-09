FROM golang:1.17-alpine as builder

RUN apk add --no-cache --update alpine-sdk~=1.0 \
    make~=4.3 \
    bash~=5.1.16 \
    protobuf-dev~=3.18.1 \
    protoc~=3.18.1

CMD ["/bin/sh"]

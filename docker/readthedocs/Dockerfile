FROM golang:1.17-alpine as builder

RUN apk add --no-cache --update alpine-sdk=1.0-r1 \
    git=2.36.0-r0 \
    make=4.3-r0 \
    bash=5.1.16-r2 \
    protobuf-dev=3.18.1-r2 \
    protoc=3.18.1-r2

RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.4.3 \
 && GO111MODULE=on go get github.com/mwitkow/go-proto-validators/...@v0.3.0 \
 && GO111MODULE=on go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2 \
 && GO111MODULE=on go get golang.org/x/tools/cmd/goimports

COPY ./build_rtd.sh /

ENTRYPOINT [ "/build_rtd.sh" ]

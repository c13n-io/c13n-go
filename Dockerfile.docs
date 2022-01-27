FROM golang:1.17-alpine as builder

# Install dependencies and install/build lnd.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    bash \
    protobuf-dev \
    protoc

RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.4.3
RUN GO111MODULE=on go get github.com/mwitkow/go-proto-validators/...@v0.3.0
RUN GO111MODULE=on go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2
RUN GO111MODULE=on go get golang.org/x/tools/cmd/goimports

COPY ./build_rtd.sh /

ENTRYPOINT [ "/build_rtd.sh" ]

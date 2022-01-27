FROM golang:1.17-alpine as builder

# Install dependencies and install/build lnd.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    bash \
    protobuf-dev \
    protoc

CMD ["/bin/sh"]

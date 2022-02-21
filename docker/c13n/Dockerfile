FROM golang:1.17-alpine as builder

# Install dependencies and install/build lnd.
RUN apk add --no-cache --update alpine-sdk \
    git \
    bash \
    make

# Copy in the local repository to build from.
COPY . /c13n

RUN cd /c13n \
    && make clean \
    && make c13n

# Start a new, final image.
FROM alpine as final

VOLUME /c13n

COPY --from=builder /c13n/c13n /bin/

ENTRYPOINT [ "c13n" ]
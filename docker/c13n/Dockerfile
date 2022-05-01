FROM golang:1.17-alpine as builder

# Install dependencies and install/build lnd.
RUN apk add --no-cache --update alpine-sdk=1.0-r1 \
    git=2.36.0-r0  \
    bash=5.1.16-r2  \
    make=4.3-r0

# Copy in the local repository to build from.
COPY . /c13n

WORKDIR /c13n

RUN make clean \
    && make c13n

# Start a new, final image.
FROM alpine:3.15.4 as final

VOLUME /c13n

COPY --from=builder /c13n/c13n /bin/

ENTRYPOINT [ "c13n" ]
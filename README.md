# c13n-go

**c13n** is a project that utilizes micro-transactions in Bitcoin's Lightning Network to transmit messages.

This repository stores the code for the project's server, which exposes a gRPC API to handle messaging requests.

### Documentation

The API documentation is hosted [here](https://docs.c13n.io/projects/api/en/latest/), while for a general introduction see [here](https://docs.c13n.io/en/latest/).

## Getting Started

These instructions will help you build the project and start the gRPC server.

### Prerequisites

#### Go 1.14

Go is an open source programming language. You can find Go for all operating systems [here](https://golang.org/dl/).

If you are using a UNIX-based system, there is a good chance you can download Go from your package manager.

Verify that Go is installed in your system.
```bash
go version
```

Also, make sure that your `$PATH` includes the `$GOBIN` folder, which is where Go places binaries from compiled packages.

### Installing

#### Download

Clone this repository.
```bash
git clone https://github.com/c13n-io/c13n-go.git
```

#### Build

Build the server.
```bash
make
```
You can also do this manually:
```bash
go build -i -v -o c13n github.com/c13n-io/c13n-go/cli
```

#### Configure

Use the `c13n.sample.yaml` file as a template to configure your app.
```bash
cp c13n.sample.yaml c13n.yaml
vim c13n.yaml
```
Note that the application requires the connectivity credentials for a Lightning daemon (`lnd`) that accepts spontaneous payments through keysend. Provide those under the `lnd` section of the configuration file.

#### TLS Certificate

A valid certificate (and key) file needs to be present if the application is to run with TLS enabled.
A self-signed certificate can be created by running
```bash
make certgen
```
or you can use a preexisting one.

#### Run

Run the server with `c13n.yaml` file as follows:
```bash
./c13n -config=c13n.yaml
```
You can start multiple instances of the server, possibly connected to different Lightning daemons, by using different configuration files for each instance.
```bash
./c13n -config=alice.yaml
./c13n -config=bob.yaml
```

### Development

#### Protocol buffer compiler

This project uses protocol buffers, which are Google's language-neutral, platform-neutral, extensible mechanism for serializing structured data. You can find the protocol buffer compiler [here](https://github.com/protocolbuffers/protobuf).

Verify that the protobuf compiler is installed in your system.
```bash
protoc --version
```

Note: The installation directory of the `protoc` binary must be in your path.

Alternatively, you can find installation instructions [here](https://grpc.io/docs/quickstart/go/).

#### Dependencies

Go can automatically install any required dependencies of the target build, however development dependencies have to be installed in a more manual manner.

You can install these dependencies using the `go get` command as outlined [below](https://dev.to/maelvls/why-is-go111module-everywhere-and-everything-about-go-modules-24k), or run `make dev-deps`.
```bash
(cd && GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.4.3)
(cd && GO111MODULE=on go get github.com/mwitkow/go-proto-validators/...@v0.3.0)
(cd && GO111MODULE=on go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2)
(cd && GO111MODULE=on go get github.com/vektra/mockery/...@v1.0.0)
(cd && GO111MODULE=on go get golang.org/x/tools/cmd/goimports)
(cd && GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0)
```
Please track any development dependencies in the above list.

#### Tests

Run the tests:
```bash
make test
```

#### Linting

c13n uses [golangci-lint](https://golangci-lint.run/) for linting, with the configuration defined in [.golangci.yml](/.golangci.yml). Linting can be triggered by running:

```bash
make lint
```

#### API Documentation

To generate the RPC documentation:
```bash
make proto-doc
```

Documentation will then be located at `docs/index.html`.

### Suggestions

#### BloomRPC

BloomRPC is an application that allows you to query gRPC services. You can download it [here](https://github.com/uw-labs/bloomrpc).

#### Polar

Polar is an application that allows developers to quickly spin up one or more Lightning Networks locally using `docker`. You can download it [here](https://github.com/jamaljsr/polar).

### Contributing

If you want to contribute to this project, either by **authoring code** or by **reporting bugs & issues**, make sure to read the [Contribution Guidelines](CONTRIBUTING.md).

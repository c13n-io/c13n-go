# c13n-backend

**c13n** is a project that utilizes micro-transactions in Bitcoin's Lightning Network to transmit messages.

This repository stores the code for the project's server, which exposes a gRPC API to handle messaging requests.

## Getting Started

These instructions will help you build the project and start the gRPC server.

### Prerequisites

##### Go 1.14

Go is an open source programming language. You can find Go for all operating systems [here](https://golang.org/dl/).

If you are using a UNIX-based system, there is a good chance you can download Go from your package manager.

Verify that Go is installed in your system.
```bash
go version
```

Also, make sure that your `$PATH` includes the `$GOBIN` folder, which is where Go places binaries from compiled packages.

### Installing

##### Private repositories

In order to access the library repository which is private at the moment, you may want to [cache](https://help.github.com/en/github/using-git/caching-your-github-password-in-git) your git credentials:
```bash
git config --global credential.helper cache
export GIT_TERMINAL_PROMPT=1
```

You may also want to set the company git base URL as [private](https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules):
```bash
export GOPRIVATE="git.programize.com"
```

##### Download

Clone this repository.
```bash
git clone https://github.com/c13n-io/c13n-backend.git
```

##### Build

Build the server.
```bash
make
```
You can also do this manually:
```bash
go build -i -v -o c13n github.com/c13n-io/c13n-backend/cli
```

##### Configure

Use the `c13n.sample.yaml` file as a template to configure your app.
```bash
cp c13n.sample.yaml c13n.yaml
vim c13n.yaml
```
Note that the application requires the connectivity credentials for an underlying Lightning node that accepts spontaneous payments through keysend. Provide those under the `lnd` section of the configuration file.

##### TLS Certificate

A valid certificate (and key) file needs to be present if the application is to run with TLS enabled.
A self-signed certificate can be created by running
```bash
make certgen
```
or you can use a preexisting one.

##### Run

Run the server with `c13n.yaml` file as:
```bash
./backend -config=c13n.yaml
```
You can start multiple instances of the server connected to different Lightning nodes, by using different configuration files for each instance.
```bash
./backend -config=alice.yaml
./backend -config=bob.yaml
```

### Development

##### Protocol buffer compiler

This project uses protocol buffers, which are Google's language-neutral, platform-neutral, extensible mechanism for serializing structured data. You can find the protocol buffer compiler [here](https://github.com/protocolbuffers/protobuf).

Verify that the protobuf compiler is installed in your system.
```bash
protoc --version
```

Note: The installation directory of the `protoc` binary must be in your path.

Alternatively, you can find installation instructions [here](https://grpc.io/docs/quickstart/go/).

##### Dependencies

Go can automatically install any required dependencies of the target build, however development dependencies have to be installed in a more manual manner.

You can install these dependencies using the `go get` command as outlined [below](https://dev.to/maelvls/why-is-go111module-everywhere-and-everything-about-go-modules-24k), or run `make dev-deps`.
```bash
(cd && GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.4.3)
(cd && GO111MODULE=on go get github.com/mwitkow/go-proto-validators/...@v0.3.0)
(cd && GO111MODULE=on go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2)
(cd && GO111MODULE=on go get golang.org/x/tools/cmd/goimports)
(cd && GO111MODULE=on go get -u github.com/mgechev/revive@v1.0.2)
(cd && GO111MODULE=on go get github.com/vektra/mockery/...@v1.0.0)
(cd && GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@v0.2.0)
```
Please track any development dependencies in the above list.

##### Tests

Run the tests:
```bash
make test
```

##### Linting

c13n uses [revive](https://github.com/mgechev/revive) for linting. Linter configuration is defined in [revive.toml](/revive.toml). To trigger linting use the following:

```bash
make lint
```

##### Documentation

###### RPC API documentation

To generate RPC API documentation:
```bash
make proto-doc
```

Documentation will then be located at `docs/index.html`.

### Suggestions

##### BloomRPC

BloomRPC is an application that allows you to query gRPC services. You can download it [here](https://github.com/uw-labs/bloomrpc).

##### Polar

Polar is an application that allows developers to quickly spin up one or more Lightning Networks locally using `docker`. You can download it [here](https://github.com/jamaljsr/polar).

### License


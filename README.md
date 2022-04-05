# c13n-go

**c13n** is a project that utilizes micro-transactions in Bitcoin's Lightning Network to transmit messages.

This repository stores the code for the project's server, which exposes a gRPC API to handle messaging requests.

### Documentation

The API documentation is hosted [here](https://docs.c13n.io/projects/api/en/latest/), while for a general introduction see [here](https://docs.c13n.io/en/latest/).

## Getting Started

These instructions will help you build the project and start the gRPC server.

### Prerequisites

#### Go 1.17

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

##### Generating database encryption key
To enable database encryption, we need to pass a file that stores the data encryption key either through the `--db-key-path` option or through the `database.key_path` configuration file parameter. The key size must be 16, 24, or 32 bytes long, and the key size determines the corresponding block size for AES encryption ,i.e. AES-128, AES-192, and AES-256, respectively.

The following command can be used to create a valid encryption key file (set count to the desired key size):

```bash
tr -dc 'a-zA-Z0-9' < /dev/urandom | dd bs=1 count=32 of=path/of/encryption/key
```

However, storing the encryption key file in the same host as the store itself defeats the purpose and is actually not more secure than leaving the database unencrypted in the first place. For this, it is highly recommended to use a password manager or vault to store the encryption key. This approach only works for filesystems that support named pipes:

```bash
# Create a named pipe
mkfifo /tmp/c13n-db-enc-key

# Fetch the password from the password manager/vault
# Using pass
pass c13n/db-enc-key > /tmp/c13n-db-enc-key &
# Using HashiCorp vault
vault kv get -field=pass c13n/db-enc-key > /tmp/c13n-db-enc-key &

# Then use the named pipe path as the database encryption key path while configuring c13n.
c13n -db-key-path=/tmp/c13n-db-enc-key
```

##### Setup configuration file
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

You can install these dependencies using the `go install` command as outlined [below](https://maelvls.dev/go111module-everywhere/), or run `make dev-deps`.
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
go install github.com/mwitkow/go-proto-validators/...@v0.3.2
go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.3.2
go install github.com/vektra/mockery/...@v1.0.0
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.1
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

#### Arc

Arc is a social wallet.
A progressive web app that natively combines messages with micropayments based on Lightning Network and <a href="https://github.com/c13n-io/c13n-go/">c13n-go</a>.
It is our reference client implementation for c13n-go. If you have a c13n node up and running, you can immediately start using Arc here: https://c13n-io.github.io/arc/

#### BloomRPC

BloomRPC is an application that allows you to query gRPC services. You can download it [here](https://github.com/uw-labs/bloomrpc).

#### Polar

Polar is an application that allows developers to quickly spin up one or more Lightning Networks locally using `docker`. You can download it [here](https://github.com/jamaljsr/polar).

### Contributing

If you want to contribute to this project, either by **authoring code** or by **reporting bugs & issues**, make sure to read the [Contribution Guidelines](CONTRIBUTING.md).

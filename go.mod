module github.com/c13n-io/c13n-go

go 1.14

require (
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/antonfisher/nested-logrus-formatter v1.1.0
	github.com/btcsuite/btcd v0.21.0-beta.0.20210513141527-ee5896bad5be
	github.com/btcsuite/btcutil v1.0.3-0.20210527170813-e2ba6805a890
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/badger v1.6.0
	github.com/go-errors/errors v1.0.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/hashicorp/go-multierror v1.0.0
	github.com/lightningnetwork/lnd v0.13.1-beta
	github.com/lightningnetwork/lnd/cert v1.0.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mwitkow/go-proto-validators v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/timshannon/badgerhold v1.0.0
	github.com/tv42/zbase32 v0.0.0-20160707012821-501572607d02
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20200212174721-66ed5ce911ce // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/macaroon.v2 v2.1.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v2 v2.2.8 // indirect
	syreclabs.com/go/faker v1.2.2
)

// Fix incompatibility of etcd go.mod package.
// See https://github.com/etcd-io/etcd/issues/11154
replace go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20201125193152-8a03d2e9614b

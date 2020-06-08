module github.com/cloudflare/goflow/v3

replace github.com/cloudflare/goflow/v3/transport => ./transport:

go 1.12

require (
	github.com/Shopify/sarama v1.22.0
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/protobuf v1.3.1
	github.com/libp2p/go-reuseport v0.0.1
	github.com/prometheus/client_golang v0.9.2
	github.com/sirupsen/logrus v1.4.1
	github.com/stretchr/testify v1.3.0
)

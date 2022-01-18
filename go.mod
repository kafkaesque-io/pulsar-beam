module github.com/kafkaesque-io/pulsar-beam

go 1.17

require (
	github.com/apache/pulsar-client-go v0.1.1-0.20200425133951-6edc8f4ef954
	github.com/ghodss/yaml v1.0.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/gops v0.3.10
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/prometheus/client_golang v1.4.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.5.0
	go.mongodb.org/mongo-driver v1.8.0
)

require (
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/yahoo/athenz v1.8.55 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// temporary pulsar client until https://github.com/apache/pulsar-client-go/pull/238 can be merged
replace github.com/apache/pulsar-client-go => github.com/zzzming/pulsar-client-go v0.0.0-20200503173951-66e589ab9740

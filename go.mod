module github.com/kafkaesque-io/pulsar-beam

go 1.13

require (
	github.com/DataDog/zstd v1.4.4 // indirect
	github.com/apache/pulsar-client-go v0.1.1-0.20200425133951-6edc8f4ef954
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/prometheus/client_golang v1.4.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.5.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.2.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
)

// temporary pulsar client until https://github.com/apache/pulsar-client-go/pull/238 can be merged
replace github.com/apache/pulsar-client-go => github.com/zzzming/pulsar-client-go v0.0.0-20200503173951-66e589ab9740

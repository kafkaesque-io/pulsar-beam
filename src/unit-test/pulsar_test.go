package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
)

func TestClientCreation(t *testing.T) {
	util.Config.DbConnectionStr = os.Getenv("PULSAR_URI")
	util.Config.DbName = os.Getenv("REST_DB_TABLE_TOPIC")
	util.Config.DbPassword = os.Getenv("PULSAR_TOKEN")
	util.Config.PbDbType = "pulsarAsDb"
	util.Config.TrustStore = os.Getenv("TrustStore")

	os.Setenv("PulsarClientOperationTimeout", "1")
	os.Setenv("PulsarClientConnectionTimeout", "1")

	_, err := pulsardriver.GetPulsarClient("pulsar://test url", "token", false)
	assert(t, err != nil, "create pulsar driver with bogus url")
	assert(t, strings.HasPrefix(err.Error(), "Could not instantiate Pulsar client: Invalid service URL:"), "match invalid service URL at pulsar client creation")

	_, err = pulsardriver.GetPulsarClient("pulsar://useast1.do.kafkaesque.io:6650", "token", true)
	errNil(t, err)

	_, err = pulsardriver.GetPulsarConsumer("pulsar://test url", "token", "topicname", "sub", "failover", "latest", "subKey")
	assert(t, err != nil, "create pulsar consumer with bogus url")

	pulsardriver.CancelPulsarConsumer("testkey") //just to bump up code coverage
}

func TestProducerObject(t *testing.T) {
	os.Setenv("PulsarClientOperationTimeout", "1")
	os.Setenv("PulsarClientConnectionTimeout", "1")

	_, err := pulsardriver.GetPulsarProducer("pulsar://test url", "tokenstring", "topicName")
	assert(t, err != nil, "create pulsar consumer with bogus url")

	// pulsardriver.SendToPulsar("pulsar://", "tokenstring", "topicName", []byte("payload"), false)

	p := pulsardriver.PulsarProducer{}
	p.UpdateTime()
	p.Close()

	c := pulsardriver.PulsarConsumer{}
	c.UpdateTime()
	c.Close()

	clt := pulsardriver.PulsarClient{}
	clt.UpdateTime()
	clt.Close()
}

package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
)

func TestClientCreation(t *testing.T) {
	util.Config.DbConnectionStr = os.Getenv("PULSAR_URI")
	util.Config.DbName = os.Getenv("REST_DB_TABLE_TOPIC")
	util.Config.DbPassword = os.Getenv("PULSAR_TOKEN")
	util.Config.PbDbType = "pulsarAsDb"
	util.Config.TrustStore = os.Getenv("TrustStore")

	_, err := pulsardriver.GetTopicDriver("pulsar://test url", "token")
	assert(t, err != nil, "create pulsar driver with bogus url")
	assert(t, strings.HasPrefix(err.Error(), "Could not instantiate Pulsar client: Invalid service URL:"), "match invalid service URL at pulsar client creation")

	_, err = pulsardriver.GetTopicDriver("pulsar://useat2.do.kafkaesque.io:6650", "token")
	errNil(t, err)

	_, err = pulsardriver.GetConsumer("pulsar://test url", "token", "topicname", "sub", "subKey", pulsar.Failover, pulsar.SubscriptionPositionLatest)
	assert(t, err != nil, "create pulsar consumer with bogus url")

	pulsardriver.CancelConsumer("testkey") //just to bump up code coverage
}

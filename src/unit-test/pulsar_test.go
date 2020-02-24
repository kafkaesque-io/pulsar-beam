package tests

import (
	"os"
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

	_, _ = pulsardriver.GetTopicDriver("pulsar://test url", "token")
	//errNil(t, err)
	//assert(t, err == nil, "create pulsar driver with bogus url")

	_, _ = pulsardriver.GetConsumer("pulsar://test url", "token", "topicname", "sub", "subKey", pulsar.Failover, pulsar.SubscriptionPositionLatest)
	//assert(t, err != nil, "create pulsar consumer with bogus url")
}

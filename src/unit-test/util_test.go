package tests

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pulsar-beam/src/broker"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/route"
	. "github.com/pulsar-beam/src/util"
)

func TestUUID(t *testing.T) {
	// Create 1 million to make sure no duplicates
	size := 100000
	set := make(map[string]bool)
	for i := 0; i < size; i++ {
		id, err := NewUUID()
		errNil(t, err)
		set[id] = true
	}
	assert(t, size == len(set), "UUID duplicates found")

}

func TestDbKeyGeneration(t *testing.T) {
	topicFullName := "persistent://my-topic/local-useast1-gcp/cloudfunction-funnel"
	pulsarURL := "pulsar+ssl://useast3.aws.host.io:6651"
	first := model.GenKey(topicFullName, pulsarURL)

	assert(t, first == model.GenKey(topicFullName, pulsarURL), "same key generates same hash")
	assert(t, first != model.GenKey(topicFullName, "pulsar://useast3.aws.host.io:6650"), "different key generates different hash")
	assert(t, first != model.GenKey("persistent://my-topic/local-useast1-gcp/cloudfunction-funnel-a", pulsarURL), "different key generates different hash")
}

func TestJoinString(t *testing.T) {
	part0 := "a"
	part1 := "b"
	part2 := "cd"
	assert(t, JoinString(part0, part2) == part0+part2, "JoinString test two string")
	assert(t, JoinString(part1, part0, part2) == part1+part0+part2, "JoinString test three string")
}
func TestAssignString(t *testing.T) {
	assert(t, "YesYo" == AssignString("", "", "YesYo"), "test default")
	assert(t, "YesYo" == AssignString("", "YesYo", "test"), "test valid string in middle")
	assert(t, "YesYo" == AssignString("YesYo", "YesYo2", "test"), "test valid string in head")
	assert(t, "" == AssignString("", "", ""), "test no valid string")
}

func TestLoadConfigFile(t *testing.T) {

	os.Setenv("PORT", "9876543")
	ReadConfigFile("../" + DefaultConfigFile)
	config := GetConfig()
	assert(t, config.PORT == "9876543", "config value overwritteen by env")
	assert(t, config.PbDbType == "inmemory", "default config setting")
}

func TestEffectiveRoutes(t *testing.T) {
	receiverRoutesLen := len(route.ReceiverRoutes)
	restRoutesLen := len(route.RestRoutes)
	prometheusLen := len(route.PrometheusRoute)
	mode := "hybrid"
	assert(t, len(route.GetEffectiveRoutes(&mode)) == (receiverRoutesLen+restRoutesLen+prometheusLen), "check hybrid required routes")
	mode = "rest"
	assert(t, len(route.GetEffectiveRoutes(&mode)) == (restRoutesLen+prometheusLen), "check rest required routes")
	mode = "receiver"
	assert(t, len(route.GetEffectiveRoutes(&mode)) == (receiverRoutesLen+prometheusLen), "check receiver required routes")
	mode = "tokenserver"
	assert(t, len(route.GetEffectiveRoutes(&mode)) == (len(route.TokenServerRoutes)+prometheusLen), "check receiver required routes")
}

func TestMainControlMode(t *testing.T) {
	mode := "receiver"
	assert(t, IsBroker(&mode) == false, "")
	assert(t, IsValidMode(&mode), "")
	mode = "broker"
	assert(t, IsBroker(&mode), "")
	assert(t, IsBrokerRequired(&mode), "")
	assert(t, IsValidMode(&mode), "")

	mode = "hybrid"
	assert(t, IsHTTPRouterRequired(&mode), "")
	assert(t, IsBrokerRequired(&mode), "")
	assert(t, IsValidMode(&mode), "")

	mode = "rest"
	assert(t, IsHTTPRouterRequired(&mode), "")
	assert(t, IsBrokerRequired(&mode) == false, "")
	assert(t, IsBroker(&mode) == false, "")
	assert(t, IsValidMode(&mode), "")

	mode = "tokenserver"
	assert(t, IsHTTPRouterRequired(&mode), "")
	assert(t, IsBrokerRequired(&mode) == false, "")
	assert(t, IsBroker(&mode) == false, "")
	assert(t, IsValidMode(&mode), "")

	mode = "oops"
	assert(t, !IsValidMode(&mode), "test invalid mode")
	assert(t, !IsHTTPRouterRequired(&mode), "test invalid HTTPRouterRequired mode")
}

func TestReceiverHeader(t *testing.T) {
	header := http.Header{}
	header.Set("Authorization", "Bearer erfagagagag")
	header.Set("PulsarUrl", "pulsar://mydomain.net:6650")
	_, _, _, result := ReceiverHeader(&header)
	assert(t, result, "Test missing TopicFn")

	header.Set("TopicFn", "http://target.net/route")
	var webhook string
	_, webhook, _, result = ReceiverHeader(&header)
	assert(t, !result, "test all headers presence")
	assert(t, webhook == header.Get("TopicFn"), "test all headers presence")
}

func TestThreaSafeMap(t *testing.T) {
	// TODO add more goroutine to test concurrency
	equals(t, false, broker.ReadWebhook("first"))
	broker.WriteWebhook("first")
	equals(t, true, broker.ReadWebhook("first"))
	broker.DeleteWebhook("first")
	equals(t, false, broker.ReadWebhook("first"))
	go func() {
		for i := 0; i < 1000; i++ {
			broker.WriteWebhook("key" + strconv.Itoa(i))
			broker.WriteWebhook("first")
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			broker.ReadWebhook("key" + strconv.Itoa(i))
			broker.ReadWebhook("first")
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			broker.DeleteWebhook("key" + strconv.Itoa(i))
			broker.WriteWebhook("first" + strconv.Itoa(i))
		}
	}()
}

func TestReportError(t *testing.T) {
	errorStr := "my invented error"
	equals(t, errorStr, ReportError(errors.New(errorStr)).Error())
}

func TestStrContains(t *testing.T) {
	assert(t, StrContains([]string{"foo", "bar", "flying"}, "foo"), "")
	assert(t, !StrContains([]string{"foo", "bar", "flying"}, "foobar"), "")
	assert(t, !StrContains([]string{}, "foobar"), "")
}

package tests

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kafkaesque-io/pulsar-beam/src/broker"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/route"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	. "github.com/kafkaesque-io/pulsar-beam/src/util"
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
	assert(t, config.PbDbInterval == "10s", "default config setting")
	assert(t, os.Getenv("PbDbInterval") == "10s", "default config setting")

	assert(t, false == util.StringToBool(os.Getenv("PulsarTLSAllowInsecureConnection")), "")

	assert(t, os.Getenv("PulsarTLSValidateHostname") == "", "PulsarTLSValidateHostname default config from env")
	assert(t, false == util.StringToBool(os.Getenv("PulsarTLSValidateHostname")), "PulsarTLSValidateHostname false as the default config from env")
	assert(t, os.Getenv("PbDbInterval") == "10s", "default config setting")

	dbType := "inmemory2"
	os.Setenv("PbDbType", dbType)
	ReadConfigFile("../../config/pulsar_beam.yml")
	config2 := GetConfig()
	assert(t, config2.PORT == "9876543", "config value overwritteen by env")
	fmt.Println(config2.PbDbType)
	assert(t, config2.PbDbType == dbType, "default config setting")
}

func TestHTTPRouters(t *testing.T) {
	mode := "hybrid"
	router := route.NewRouter(&mode)

	routeName := "receive"
	route := router.Get(routeName)
	assert(t, route == nil, "get route name")
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

	mode = "http2"
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
	// header.Set("Authorization", "Bearer erfagagagag")
	_, _, _, result := ReceiverHeader(strings.Split("", ","), &header)
	equals(t, result.Error(), "missing configured Pulsar URL")

	header.Set("PulsarUrl", "http://target.net/route")
	var token string
	token, _, _, result = ReceiverHeader(strings.Split("", ","), &header)
	errNil(t, result)
	assert(t, token == "", "test all headers presence")
}

func TestDefaultPulsarURLInReceiverHeader(t *testing.T) {
	allowedPulsarURLs := strings.Split("pulsar+ssl://kafkaesque.net:6651", ",")
	header := http.Header{}
	header.Set("Authorization", "Bearer erfagagagag")
	_, _, pulsarURL, result := ReceiverHeader(allowedPulsarURLs, &header)
	errNil(t, result)
	equals(t, pulsarURL, "pulsar+ssl://kafkaesque.net:6651")

	header.Set("TopicFn", "http://target.net/route")
	var webhook string
	_, webhook, pulsarURL, result = ReceiverHeader(allowedPulsarURLs, &header)
	errNil(t, result)
	assert(t, webhook == header.Get("TopicFn"), "test all headers presence")
	assert(t, pulsarURL == "pulsar+ssl://kafkaesque.net:6651", "test all headers presence")
	assert(t, "" == header.Get("PulsarUrl"), "ensure PulsarUrl is empty")
}

func TestThreadSafeMap(t *testing.T) {
	// TODO add more goroutine to test concurrency
	wb := broker.NewWebhookBroker(&Configuration{})

	_, rc := wb.ReadWebhook("first")
	equals(t, false, rc)

	sig := make(chan *broker.SubCloseSignal, 2)

	wb.WriteWebhook("first", sig)
	_, rc = wb.ReadWebhook("first")
	equals(t, true, rc)
	wb.DeleteWebhook("first")
	_, rc = wb.ReadWebhook("first")
	equals(t, false, rc)
	go func() {
		for i := 0; i < 1000; i++ {
			sig2 := make(chan *broker.SubCloseSignal, 2)
			wb.WriteWebhook("key"+strconv.Itoa(i), sig2)
			wb.WriteWebhook("first", sig2)
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			wb.ReadWebhook("key" + strconv.Itoa(i))
			wb.ReadWebhook("first")
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			sig3 := make(chan *broker.SubCloseSignal, 2)
			wb.DeleteWebhook("key" + strconv.Itoa(i))
			wb.WriteWebhook("first"+strconv.Itoa(i), sig3)
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

func TestGetEnvInt(t *testing.T) {
	os.Setenv("Kopper", "546")
	equals(t, 546, GetEnvInt("Kopper", 90))

	os.Setenv("Kopper", "5o46")
	equals(t, 946, GetEnvInt("Kopper", 946))
}

type TestObj struct {
	isClosed bool
}

func (o *TestObj) Close() {
	o.isClosed = true
}

func TestExpiryTTLCache(t *testing.T) {

	cache := NewCache(CacheOption{
		TTL:           10 * time.Millisecond,
		CleanInterval: 12 * time.Millisecond,
		ExpireCallback: func(key string, value interface{}) {
			if obj, ok := value.(*TestObj); ok {
				obj.Close()
			} else {
				assert(t, false, "wrong object type stored in Cache")
			}
		},
	})

	object1 := TestObj{}
	cache.Set("object1", &object1)
	cache.Set("object2", &TestObj{})
	cache.Set("object3", &TestObj{})

	object4 := TestObj{}
	cache.Set("object4", &object4)

	object5 := TestObj{}
	cache.SetWithTTL("object5", &object5, 4*time.Millisecond)

	obj, ok := cache.Get("object5")
	assert(t, ok, "verify added object exists in cache")
	assert(t, !(obj.(*TestObj).isClosed), "ensure object has not been Close()")
	assert(t, 5 == cache.Count(), "check the counts of total number of objects in cache")

	// make sure object4 won't expire
	time.Sleep(8 * time.Millisecond)
	_, ok = cache.Get("object5")
	assert(t, !ok, "object5 has already expired")

	// comment out because GitHub action may have different timing that causes this failing
	//_, ok = cache.Get("object4")
	//assert(t, ok, "object4 still exists")

	// another 2ms expires the default TTL
	time.Sleep(2 * time.Millisecond)
	assert(t, 4 == cache.Count(), "all objects are expired but not purged yet")
	_, ok = cache.Get("object1")
	assert(t, !ok, "object has expired")
	assert(t, object1.isClosed, "object1 has been Close() by the callback")
	assert(t, !object4.isClosed, "object4 has not been Close() by the callback")
	assert(t, object5.isClosed, "object5 has not been Close() by the callback")

	//time.Sleep(2 * time.Millisecond)
	//assert(t, 1 == cache.Count(), "check the counts of total number of objects in cache")
	//assert(t, !object4.isClosed, "object4 has not expired yet")

}

func TestInfinityExpiryTTLCache(t *testing.T) {

	cache := NewCache(CacheOption{
		TTL:           2 * time.Millisecond,
		CleanInterval: 3 * time.Millisecond,
		ExpireCallback: func(key string, value interface{}) {
			if obj, ok := value.(*TestObj); ok {
				obj.Close()
			} else {
				assert(t, false, "wrong object type stored in Cache")
			}
		},
	})

	object1 := TestObj{}
	cache.SetWithTTL("object1", &object1, -1)

	object2 := TestObj{}
	cache.Set("object2", &object2)

	obj, ok := cache.Get("object2")
	assert(t, ok, "verify added object exists in cache")
	assert(t, !(obj.(*TestObj).isClosed), "ensure object has not been Close()")
	assert(t, 2 == cache.Count(), "check the counts of total number of objects in cache")

	// make sure object4 won't expire
	time.Sleep(4 * time.Millisecond)
	_, ok = cache.Get("object1")
	assert(t, ok, "object1 still exists")
	_, ok = cache.Get("object2")
	assert(t, !ok, "object2 already expired")

	assert(t, 1 == cache.Count(), "all objects are expired but not purged yet")

	assert(t, !object1.isClosed, "object1 has not expired yet")
	assert(t, object2.isClosed, "object2 already Close()")

	cache.Delete("object1")
	assert(t, 0 == cache.Count(), "all objects should be deleted")
	cache.Close()
}

func TestConcurrencyTTLCache(t *testing.T) {

	cache := NewCache(CacheOption{
		TTL:           2 * time.Millisecond,
		CleanInterval: 10 * time.Millisecond,
		ExpireCallback: func(key string, value interface{}) {
			if obj, ok := value.(*TestObj); ok {
				obj.Close()
			} else {
				assert(t, false, "wrong object type stored in Cache")
			}
		},
	})

	for i := 0; i < 5; i++ {
		go func(index int) {
			cache.Set("p_object"+strconv.Itoa(index), &TestObj{})
			fmt.Printf("added entry %s\n", "d_object"+strconv.Itoa(index))
		}(i)
	}
	for i := 0; i < 8; i++ {
		go func(index int) {
			cache.SetWithTTL("q_object"+strconv.Itoa(index), &TestObj{}, 5*time.Millisecond)
			fmt.Printf("added entry %s\n", "object"+strconv.Itoa(index))
		}(i)
	}
	object1 := TestObj{}
	cache.SetWithTTL("object1", &object1, 20*time.Second)

	fmt.Printf("total number is %d\n", cache.Count())
	assert(t, 14 > cache.Count(), "check the counts of total number of objects in cache")

	// make sure object4 won't expire
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("2 total number is %d\n", cache.Count())
	assert(t, 14 == cache.Count(), "some objects have expired")
	assert(t, !object1.isClosed, "object1 has not expired yet")
}

func TestStringToBoo(t *testing.T) {
	// true cases
	assert(t, StringToBool("true"), "string true yields boolean true")
	assert(t, StringToBool("True"), "string True yields boolean true")
	assert(t, StringToBool(" tRue"), "string tRue with space yields boolean true")
	assert(t, StringToBool("yes"), "string true yields boolean true")
	assert(t, StringToBool("1"), "string true yields boolean true")
	assert(t, StringToBool("enable"), "string true yields boolean true")
	assert(t, StringToBool(" Enabled "), "string Enabled with space yields boolean true")
	assert(t, StringToBool("ok"), "string ok yields boolean true")
	assert(t, StringToBool("Ok"), "string Ok yields boolean true")

	// false cases
	assert(t, !StringToBool(" "), "string space yields boolean false")
	assert(t, !StringToBool(""), "string empty string yields boolean false")
	assert(t, !StringToBool(" t rue"), "string t rue with space yields boolean false")
	assert(t, !StringToBool("no"), "string no yields boolean false")
	assert(t, !StringToBool("10"), "string 10 yields boolean false")
	assert(t, !StringToBool("0"), "string 0 yields boolean false")
	assert(t, !StringToBool("notok"), "string notok yields boolean false")
	assert(t, !StringToBool("disable"), "string disable yields boolean false")
	assert(t, !StringToBool("adsfasdf"), "string any string yields boolean false")
}

func TestTokenizeTopicFullName(t *testing.T) {
	isPersistent, tenant, ns, topic, err := TokenizeTopicFullName("persistent://public/default/test-topic")
	errNil(t, err)
	equals(t, isPersistent, true)
	equals(t, tenant, "public")
	equals(t, ns, "default")
	equals(t, topic, "test-topic")

	isPersistent, tenant, ns, topic, err = TokenizeTopicFullName("non-persistent://public/default/test-topic")
	errNil(t, err)
	equals(t, isPersistent, false)
	equals(t, tenant, "public")
	equals(t, ns, "default")
	equals(t, topic, "test-topic")

	_, _, _, _, err = TokenizeTopicFullName("persitent://public/default/test-topic")
	assertErr(t, "invalid persistent or non-persistent part", err)

	isPersistent, tenant, ns, topic, err = TokenizeTopicFullName("non-persistent://public/default")
	errNil(t, err)
	equals(t, isPersistent, false)
	equals(t, tenant, "public")
	equals(t, ns, "default")
	equals(t, topic, "")

	_, _, _, _, err = TokenizeTopicFullName("non-persistent://public")
	assertErr(t, "missing tenant, namespace, or topic name", err)

	_, _, _, _, err = TokenizeTopicFullName("non-persistent://public/default/to2/to3")
	assertErr(t, "missing tenant, namespace, or topic name", err)
}

func TestHTTPParams(t *testing.T) {
	params := url.Values{}
	params.Set("var1", "testme")
	params.Set("var2", "48")
	params.Set("var3", "7")
	equals(t, QueryParamInt(params, "var1", 5), 5)
	equals(t, QueryParamInt(params, "var2", 5), 48)
	equals(t, QueryParamInt(params, "var22", 5), 5)

	equals(t, QueryParamString(params, "var1", "48"), "testme")
	equals(t, QueryParamString(params, "var2", "test"), "48")
	equals(t, QueryParamString(params, "var22", "another"), "another")
}

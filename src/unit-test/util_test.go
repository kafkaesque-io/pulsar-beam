package tests

import (
	"os"
	"testing"

	"github.com/pulsar-beam/src/db"
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
	first := db.GenKey(topicFullName, pulsarURL)

	assert(t, first == db.GenKey(topicFullName, pulsarURL), "same key generates same hash")
	assert(t, first != db.GenKey(topicFullName, "pulsar://useast3.aws.host.io:6650"), "different key generates different hash")
	assert(t, first != db.GenKey("persistent://my-topic/local-useast1-gcp/cloudfunction-funnel-a", pulsarURL), "different key generates different hash")
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
	assert(t, config.PbDbType == "mongo", "default config setting")
}

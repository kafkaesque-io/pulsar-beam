package util

import (
	"encoding/json"
	"log"
	"os"
	"reflect"

	"github.com/pulsar-beam/src/icrypto"
)

// DefaultConfigFile - default config file
// it can be overwritten by env variable PULSAR_BEAM_CONFIG
const DefaultConfigFile = "../config/pulsar_beam.json"

// Configuration - this server's configuration
type Configuration struct {
	PORT    string `json:"PORT"`
	CLUSTER string `json:"CLUSTER"`
	User    string `json:"User"`
	Pass    string `json:"Pass"`

	// DbName is the database name in mongo or topic name
	DbName string `json:"DbName"`

	// DbPassword is either password or token for the database
	DbPassword string `json:"DbPassword"`

	// DbConnectionStr can be mongo url or pulsar url
	DbConnectionStr string `json:"DbConnectionStr"`

	// PbDbType is the database type mongo or pulsar
	PbDbType         string `json:"PbDbType"`
	PulsarPublicKey  string `json:"PulsarPublicKey"`
	PulsarPrivateKey string `json:"PulsarPrivateKey"`
	SuperRoles       string `json:"SuperRoles"`

	// Webhook consumers pool checked interval to stop deleted consumers and start new ones
	// default value 180s
	PbDbInterval string `json:"PbDbInterval"`

	// Pulsar CA certificate key store
	TrustStore string `json:"TrustStore"`
}

// Config - this server's configuration instance
var Config Configuration

// JWTAuth is the RSA key pair for sign and verify JWT
var JWTAuth *icrypto.RSAKeyPair

// Init initializes configuration
func Init() {
	configFile := AssignString(os.Getenv("PULSAR_BEAM_CONFIG"), DefaultConfigFile)
	log.Printf("Configuration built from file - %s", configFile)
	ReadConfigFile(configFile)

	JWTAuth = icrypto.NewRSAKeyPair(Config.PulsarPrivateKey, Config.PulsarPublicKey)
}

// ReadConfigFile reads configuration file.
func ReadConfigFile(configFile string) {

	//filename is the path to the json config file
	file, err := os.Open(configFile)
	if err != nil {
		log.Printf("failed to load configuraiton file %s", configFile)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		panic(err)
	}

	// Next section allows env variable overwrites config file value
	fields := reflect.TypeOf(Config)
	// pointer to struct
	values := reflect.ValueOf(&Config)
	// struct
	st := values.Elem()
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i).Name
		envV := os.Getenv(field)
		if len(envV) > 0 {
			f := st.FieldByName(field)
			if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
				f.SetString(envV)
			}
		}
	}

	log.Println(Config.PORT, Config.PbDbType, Config.PbDbInterval)
}

//GetConfig returns a reference to the Configuration
func GetConfig() *Configuration {
	return &Config
}

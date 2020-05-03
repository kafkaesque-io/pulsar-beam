package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"unicode"

	"github.com/ghodss/yaml"
	"github.com/kafkaesque-io/pulsar-beam/src/icrypto"
	log "github.com/sirupsen/logrus"
)

// DefaultConfigFile - default config file
// it can be overwritten by env variable PULSAR_BEAM_CONFIG
const DefaultConfigFile = "../config/pulsar_beam.yml"

// Configuration - this server's configuration
type Configuration struct {
	PORT     string `json:"PORT"`
	CLUSTER  string `json:"CLUSTER"`
	LogLevel string `json:"LogLevel"`
	User     string `json:"User"`
	Pass     string `json:"Pass"`

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
	PulsarBrokerURL  string `json:"PulsarBrokerURL"`

	// Configure whether the Pulsar client accept untrusted TLS certificate from broker (default: false)
	// Set to `true` to enable
	PulsarTLSAllowInsecureConnection string `json:"PulsarTLSAllowInsecureConnection"`

	// Configure whether the Pulsar client verify the validity of the host name from broker (default: false)
	// Set to `true` to enable
	PulsarTLSValidateHostname string `json:"PulsarTLSValidateHostname"`

	// Webhook consumers pool checked interval to stop deleted consumers and start new ones
	// default value 180s
	PbDbInterval string `json:"PbDbInterval"`

	// Pulsar CA certificate key store
	TrustStore string `json:"TrustStore"`

	// TLS
	CertFile string `json:"CertFile"`
	KeyFile  string `json:"KeyFile"`

	// PulsarClusters enforce that Beam only can connect to the specified clusters
	// It is a comma separated pulsar URL string, so it can be a list of clusters
	PulsarClusters string `json:"PulsarClusters"`

	// HTTPAuthImpl specifies the jwt authen and authorization algorithm, `noauth` to skip authentication
	HTTPAuthImpl string `json:"HTTPAuthImpl"`
}

var (
	// AllowedPulsarURLs specifies a list of allowed pulsar URL/cluster
	AllowedPulsarURLs []string

	// SuperRoles are admin level users for jwt authorization
	SuperRoles []string

	// Config - this server's configuration instance
	Config Configuration

	// JWTAuth is the RSA key pair for sign and verify JWT
	JWTAuth *icrypto.RSAKeyPair

	// L is the logger
	L *log.Logger
)

// Init initializes configuration
func Init() {
	configFile := AssignString(os.Getenv("PULSAR_BEAM_CONFIG"), DefaultConfigFile)
	ReadConfigFile(configFile)

	log.SetLevel(logLevel(Config.LogLevel))

	log.Warnf("Configuration built from file - %s", configFile)
	JWTAuth = icrypto.NewRSAKeyPair(Config.PulsarPrivateKey, Config.PulsarPublicKey)
}

// ReadConfigFile reads configuration file.
func ReadConfigFile(configFile string) {
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("failed to load configuration file %s", configFile)
		panic(err)
	}

	if hasJSONPrefix(fileBytes) {
		err = json.Unmarshal(fileBytes, &Config)
		if err != nil {
			panic(err)
		}
	} else {
		err = yaml.Unmarshal(fileBytes, &Config)
		if err != nil {
			panic(err)
		}
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
				f.SetString(strings.TrimSuffix(envV, "\n")) // ensure no \n at the end of line that was introduced by loading k8s secrete file
			}
		}
	}

	clusterStr := AssignString(Config.PulsarClusters, "")
	AllowedPulsarURLs = strings.Split(clusterStr, ",")
	if Config.PulsarBrokerURL != "" {
		AllowedPulsarURLs = append([]string{Config.PulsarBrokerURL}, AllowedPulsarURLs...)
	}

	superRoleStr := AssignString(Config.SuperRoles, "superuser")
	SuperRoles = strings.Split(superRoleStr, ",")

	fmt.Printf("port %s, PbDbType %s, DbRefreshInterval %s, TrustStore %s, DbName %s, DbConnectString %s\n",
		Config.PORT, Config.PbDbType, Config.PbDbInterval, Config.TrustStore, Config.DbName, Config.DbConnectionStr)
	fmt.Printf("PublicKey %s, PrivateKey %s\n",
		Config.PulsarPublicKey, Config.PulsarPrivateKey)
	fmt.Printf("PulsarBrokerURL %s, AllowedPulsarURLs %v,PulsarTLSAllowInsecureConnection %s,PulsarTLSValidateHostname %s\n",
		Config.PulsarBrokerURL, AllowedPulsarURLs, Config.PulsarTLSAllowInsecureConnection, Config.PulsarTLSValidateHostname)
}

//GetConfig returns a reference to the Configuration
func GetConfig() *Configuration {
	return &Config
}

func logLevel(level string) log.Level {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case "debug":
		return log.DebugLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}

var jsonPrefix = []byte("{")

func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

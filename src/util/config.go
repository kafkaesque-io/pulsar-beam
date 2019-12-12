package util

import (
	"encoding/json"
	"log"
	"os"
)

// ConfigurationFile -
const ConfigurationFile = "./.pconfig"

// Configuration - this server's configuration
type Configuration struct {
	Port               int    `json:"port"`
	Host               string `json:"Host"`
	User               string `json:"User"`
	Pass               string `json:"Pass"`
	DbName             string `json:"DbName"`
	GOauthClientID     string `json:"gOauthClientID"`
	GOauthClientSecret string `json:"gOauthClientSecret"`

	AuthAudience string `json:"authAudience"`
	AuthDomain   string `json:"authDomain"`
}

// Config - this server's configuration instance
var Config Configuration

func init() {
	// fromFile()
	log.Printf("Configuration built - ")

}

func fromFile() {
	//filename is the path to the json config file
	file, err := os.Open(ConfigurationFile)
	if err != nil {
		log.Printf("failed to load configuraiton file %s", ConfigurationFile)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		panic(err)
	}
	log.Printf("Configuration built from the file - ")
}

//GetConfig returns a reference to the Configuration
func GetConfig() *Configuration {
	return &Config
}

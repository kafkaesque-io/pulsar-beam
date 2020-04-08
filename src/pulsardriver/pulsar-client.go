package pulsardriver

import (
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	log "github.com/sirupsen/logrus"
)

// ClientCache caches a list Pulsar clients
var ClientCache = make(map[string]*PulsarClient)

// clientSync protects the ClientCache access
var clientSync = &sync.RWMutex{}

var (
	clientOpsTimeout     = util.GetEnvInt("PulsarClientOperationTimeout", 30)
	clientConnectTimeout = util.GetEnvInt("PulsarClientConnectionTimeout", 30)
)

// GetPulsarClient gets a Pulsar client object
func GetPulsarClient(pulsarURL, pulsarToken string, reset bool) (pulsar.Client, error) {
	key := pulsarURL + pulsarToken
	clientSync.Lock()
	driver, ok := ClientCache[key]
	if !ok {
		driver = &PulsarClient{}
		driver.createdAt = time.Now()
		driver.pulsarURL = pulsarURL
		driver.token = pulsarToken
		ClientCache[key] = driver
	}
	clientSync.Unlock()
	if reset {
		return driver.Reconnect()
	}
	return driver.GetClient(pulsarURL, pulsarToken)

}

// PulsarClient encapsulates the Pulsar Client object
type PulsarClient struct {
	client    pulsar.Client
	pulsarURL string
	token     string
	createdAt time.Time
	lastUsed  time.Time
	sync.Mutex
}

// GetClient acquires a new pulsar client
func (c *PulsarClient) GetClient(url, tokenStr string) (pulsar.Client, error) {
	c.Lock()
	defer c.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	clientOpt := pulsar.ClientOptions{
		URL:               url,
		OperationTimeout:  time.Duration(clientOpsTimeout) * time.Second,
		ConnectionTimeout: time.Duration(clientConnectTimeout) * time.Second,
	}

	if tokenStr != "" {
		clientOpt.Authentication = pulsar.NewAuthenticationToken(tokenStr)
	}

	// TODO: add code to tell CentOS or Ubuntu
	trustStore := util.AssignString(util.GetConfig().TrustStore, "") //"/etc/ssl/certs/ca-bundle.crt"
	if trustStore != "" {
		clientOpt.TLSTrustCertsFilePath = trustStore
	}

	driver, err := pulsar.NewClient(clientOpt)

	if err != nil {
		log.Errorf("failed instantiate pulsar client %v", err)
		return nil, fmt.Errorf("Could not instantiate Pulsar client: %v", err)
	}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("pulsar client url %s\n token %s", url, tokenStr)
	}

	c.client = driver
	return driver, nil
}

// UpdateTime updates all time stamps in the object
func (c *PulsarClient) UpdateTime() {
	c.lastUsed = time.Now()
}

// Close closes the Pulsar client
func (c *PulsarClient) Close() {
	c.Lock()
	defer c.Unlock()
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}

// Reconnect closes the current connection and reconnects again
func (c *PulsarClient) Reconnect() (pulsar.Client, error) {
	c.Close()
	return c.GetClient(c.pulsarURL, c.token)
}

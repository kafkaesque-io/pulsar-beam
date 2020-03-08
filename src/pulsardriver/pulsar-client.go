package pulsardriver

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/util"
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

	// TODO: add code to tell CentOS or Ubuntu
	trustStore := util.AssignString(util.GetConfig().TrustStore, "/etc/ssl/certs/ca-bundle.crt")
	token := pulsar.NewAuthenticationToken(tokenStr)

	driver, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                   url,
		Authentication:        token,
		TLSTrustCertsFilePath: trustStore,
		OperationTimeout:      time.Duration(clientOpsTimeout) * time.Second,
		ConnectionTimeout:     time.Duration(clientConnectTimeout) * time.Second,
	})

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("Could not instantiate Pulsar client: %v", err)
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

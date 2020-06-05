package util

// This is a TLS certificate loader that detects the change of key and cert.
// One use case is to reload letsencrypt certificates
// Example to generate test certificate and key files
// openssl req -newkey rsa:2048 -nodes -keyout domain.key -x509 -days 365 -out domain.crt

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var cert atomic.Value

type updatedChann struct{}

func loadCert(certFile, keyFile string) error {
	c, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("failed to LoadX509KeyPair %v", err)
		return err
	}

	log.Println("successfully load certs and keys")
	cert.Store(c)
	return nil
}

type fileState struct {
	info os.FileInfo
	err  error
}

func watchFile(filePath string, updated chan *updatedChann) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.Tick(1 * time.Second):
			stat, err := os.Stat(filePath)
			if err == nil {
				if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
					initialStat = stat
					updated <- &updatedChann{}
				}
			}
		}
	}
}

// ListenAndServeTLS listens HTTP with TLS option just like the default http.ListenAndServeTLS
// in addition it also watches certificate and key file changes and reloads them if necessary
func ListenAndServeTLS(address, certFile, keyFile string, handler http.Handler) error {
	if len(certFile) > 1 && len(keyFile) > 1 {
		return listenAndServeTLS(address, certFile, keyFile, handler)
	}
	return http.ListenAndServe(address, handler)
}

func listenAndServeTLS(address, certFile, keyFile string, handler http.Handler) error {
	log.Printf("load certs %s and key files %s\n", certFile, keyFile)
	if err := loadCert(certFile, keyFile); err != nil {
		return err
	}

	go func(cert, key string) {
		certMonitorChan := make(chan *updatedChann, 1)
		keyMonitorChan := make(chan *updatedChann, 1)

		go watchFile(certFile, certMonitorChan)
		go watchFile(keyFile, keyMonitorChan)

		var certUpdated, keyUpdated bool
		// only update X509 key pair when both cert and key files are updated
		for {
			select {
			case <-certMonitorChan:
				certUpdated = true
				if certUpdated == keyUpdated {
					certUpdated = false
					keyUpdated = false
					loadCert(certFile, keyFile)
				}
			case <-keyMonitorChan:
				keyUpdated = true
				if certUpdated == keyUpdated {
					certUpdated = false
					keyUpdated = false
					loadCert(certFile, keyFile)
				}
			}
		}
	}(certFile, keyFile)
	// Create tlsConfig that uses a custom GetCertificate method
	// Defined by GetCertificate func at  https://golang.org/pkg/crypto/tls/
	tlsConfig := tls.Config{
		GetCertificate: func(i *tls.ClientHelloInfo) (*tls.Certificate, error) {
			c, ok := cert.Load().(tls.Certificate)
			if !ok {
				return nil, fmt.Errorf("Unable to load cert: %+v", c)
			}

			return &c, nil
		},
	}

	// listen on the port with TLS listener
	l, err := tls.Listen("tcp", address, &tlsConfig)
	if err != nil {
		return err
	}

	return http.Serve(l, handler)
}

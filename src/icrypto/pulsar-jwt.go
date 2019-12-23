package icrypto

// This is JWT sign/verify with the same key algo used in Pulsar.

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// RSAKeyPair for JWT token sign and verification
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

const (
	tokenDuration = 24
	expireOffset  = 3600
)

var jwtRsaKeys *RSAKeyPair = nil

// NewRSAKeyPair creates a pair of RSA key for JWT token sign and verification
func NewRSAKeyPair(privateKeyPath, publicKeyPath string) *RSAKeyPair {
	if jwtRsaKeys == nil {
		jwtRsaKeys = &RSAKeyPair{
			PrivateKey: getPrivateKey(privateKeyPath),
			PublicKey:  getPublicKey(publicKeyPath),
		}
	}

	return jwtRsaKeys
}

// GenerateToken generates token with user defined subject
func (keys *RSAKeyPair) GenerateToken(userSubject string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = jwt.MapClaims{
		// "exp": time.Now().Add(time.Hour * time.Duration(24)).Unix(),
		// "iat": time.Now().Unix(),
		"sub": userSubject,
	}
	tokenString, err := token.SignedString(keys.PrivateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// DecodeToken decodes a token string
func (keys *RSAKeyPair) DecodeToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtRsaKeys.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if token.Valid {
		return token, nil
	}

	return nil, errors.New("invalid token")
}

// VerifyTokenSubject verifies a token string based on required matching subject
func (keys *RSAKeyPair) VerifyTokenSubject(tokenStr, subject string) (bool, error) {
	token, err := keys.DecodeToken(tokenStr)

	if err != nil {
		return false, err
	}

	claims := token.Claims.(jwt.MapClaims)

	if subject == claims["sub"] {
		return true, nil
	}

	return false, errors.New("incorrect sub")
}

func (keys *RSAKeyPair) getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

func getPrivateKey(privateKeyPath string) *rsa.PrivateKey {
	privateKeyFile, err := os.Open(privateKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyFile.Close()

	// PKCS8 comply with Pulsar Java generated private key
	key, err := x509.ParsePKCS8PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	privateKeyImported, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("expected key to be of type *ecdsa.PrivateKey, but actual was %T", key)
	}

	return privateKeyImported
}

func getPublicKey(publicKeyPath string) *rsa.PublicKey {
	publicKeyFile, err := os.Open(publicKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}

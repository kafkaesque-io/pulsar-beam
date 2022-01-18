package icrypto

// This is JWT sign/verify with the same key algo used in Pulsar.

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
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

var jwtRsaKeys *RSAKeyPair

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

//TODO: support multiple subjects in claims

// GetTokenSubject gets the subjects from a token
func (keys *RSAKeyPair) GetTokenSubject(tokenStr string) (string, error) {
	token, err := keys.DecodeToken(tokenStr)
	if err != nil {
		return "", err
	}
	claims := token.Claims.(jwt.MapClaims)
	subjects, ok := claims["sub"]
	if ok {
		return subjects.(string), nil
	}
	return "", errors.New("missing subjects")
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

// GetTokenRemainingValidity is the remaining seconds before token expires
func (keys *RSAKeyPair) GetTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

// supports pk12 jks binary format
func readPK12(file string) ([]byte, error) {
	osFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(osFile, 4)

	return ioutil.ReadAll(reader)
}

// decode PEM format to array of bytes
func decodePEM(pemFilePath string) ([]byte, error) {
	keyFile, err := os.Open(pemFilePath)
	defer keyFile.Close()
	if err != nil {
		return nil, err
	}

	pemfileinfo, _ := keyFile.Stat()
	pembytes := make([]byte, pemfileinfo.Size())

	buffer := bufio.NewReader(keyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))
	return data.Bytes, err
}

func parseX509PKCS8PrivateKey(data []byte) *rsa.PrivateKey {
	key, err := x509.ParsePKCS8PrivateKey(data)

	if err != nil {
		panic(err)
	}

	rsaPrivate, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("expected key to be of type *ecdsa.PrivateKey, but actual was %T", key)
	}

	return rsaPrivate
}

func parseX509PKIXPublicKey(data []byte) *rsa.PublicKey {
	publicKeyImported, err := x509.ParsePKIXPublicKey(data)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)
	if !ok {
		panic(err)
	}

	return rsaPub
}

// Since we support PEM And binary fomat of PKCS12/X509 keys,
// this function tries to determine which format
func fileFormat(file string) (string, error) {
	osFile, err := os.Open(file)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReaderSize(osFile, 4)
	// attempt to guess based on first 4 bytes of input
	data, err := reader.Peek(4)
	if err != nil {
		return "", err
	}

	magic := binary.BigEndian.Uint32(data)
	if magic == 0x2D2D2D2D || magic == 0x434f4e4e {
		// Starts with '----' or 'CONN' (what s_client prints...)
		return "PEM", nil
	}
	if magic&0xFFFF0000 == 0x30820000 {
		// Looks like the input is DER-encoded, so it's either PKCS12 or X.509.
		if magic&0x0000FF00 == 0x0300 {
			// Probably X.509
			return "DER", nil
		}
		return "PKCS12", nil
	}

	return "", errors.New("undermined format")
}

func getDataFromKeyFile(file string) ([]byte, error) {
	format, err := fileFormat(file)
	if err != nil {
		return nil, err
	}

	switch format {
	case "PEM":
		return decodePEM(file)
	case "PKCS12":
		return readPK12(file)
	default:
		return nil, errors.New("unsupported format")
	}
}

func getPrivateKey(file string) *rsa.PrivateKey {
	data, err := getDataFromKeyFile(file)
	if err != nil {
		log.Fatalf("failed to load private key %v", err)
	}

	return parseX509PKCS8PrivateKey(data)
}

func getPublicKey(file string) *rsa.PublicKey {
	data, err := getDataFromKeyFile(file)
	if err != nil {
		log.Fatalf("failed to load public key %v", err)
	}

	return parseX509PKIXPublicKey(data)
}

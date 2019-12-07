package icrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io"
)

const (
	// AesAlgo enum for AES
	AesAlgo int = iota
	// RsaAlgo enum for RSA
	RsaAlgo
)

// AsymKeys asymmetric keys
type AsymKeys interface {
	getPrivateKey() ([]byte, error)
	getPublicKey() ([]byte, error)
}

// Encrypto encryption
type Encrypto interface {
	encrypt(plaintext []byte, key []byte) ([]byte, error)
	encryptWithDefaultKey(plaintext []byte) ([]byte, error)
}

// Decrypto decryption
type Decrypto interface {
	decrypt(ciphertext []byte, key []byte) ([]byte, error)
	decryptWithDefaultKey(plaintext []byte) ([]byte, error)
}

// AES struct implementation including encryption and decryption
type AES struct {
	DefaultSalt string
}

// Encrypt encrypts with asymmetric key
func (a *AES) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// EncryptWithDefaultKey encrypts with a default key
func (a *AES) EncryptWithDefaultKey(plaintext []byte) ([]byte, error) {
	defaultKey := []byte(a.DefaultSalt)
	return a.Encrypt(plaintext, defaultKey)
}

// Decrypt decrypts with asymmetric key
func (a *AES) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// DecryptWithDefaultKey decrypts with a default key
func (a *AES) DecryptWithDefaultKey(plaintext []byte) ([]byte, error) {
	defaultKey := []byte(a.DefaultSalt)
	return a.Decrypt(plaintext, defaultKey)
}

//RSA struct implementation including encryption and decryption
type RSA struct {
	MyPrivateKey *rsa.PrivateKey
	MyPublicKey  *rsa.PublicKey
}

// NewRSAWithKeys with private and public keys
func NewRSAWithKeys(priv []byte, pub []byte) (RSA, error) {
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	var err error
	if priv != nil {
		//pemblock, _ := pem.Decode(priv)
		privKey, err = x509.ParsePKCS1PrivateKey(priv)
		if err != nil {
			return RSA{}, err
		}
	}
	if pub != nil {
		//pemblock2, _ := pem.Decode(pub)
		pubKey, err = x509.ParsePKCS1PublicKey(pub)
		if err != nil {
			return RSA{}, err
		}
	}
	return RSA{MyPrivateKey: privKey, MyPublicKey: pubKey}, nil
}

// NewRSA creates RSA keys pair
func NewRSA() (RSA, error) {
	keys := RSA{}
	var err error

	keys.MyPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return RSA{}, err
	}

	keys.MyPublicKey = &(keys.MyPrivateKey.PublicKey)
	return keys, nil
}

// GetPublicKey gets the public RSA key
func (a *RSA) GetPublicKey() ([]byte, error) {
	key := x509.MarshalPKCS1PublicKey(a.MyPublicKey)
	return key, nil
}

// GetPrivateKey gets the private RSA key
func (a *RSA) GetPrivateKey() ([]byte, error) {
	key := x509.MarshalPKCS1PrivateKey(a.MyPrivateKey)
	return key, nil
}

// Encrypt encrypts with RSA key
func (a *RSA) Encrypt(plaintext []byte, key []byte) ([]byte, error) {

	return nil, errors.New("unsupported")
}

// EncryptWithDefaultKey encrypts with a default RSA key
func (a *RSA) EncryptWithDefaultKey(plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, a.MyPublicKey, plaintext)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// Decrypt decrypts with RSA key
func (a *RSA) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {

	return nil, errors.New("unsupported")
}

// DecryptWithDefaultKey decrypts with a default RSA key
func (a *RSA) DecryptWithDefaultKey(ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, a.MyPrivateKey, ciphertext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

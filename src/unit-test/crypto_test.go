package tests

import (
	"testing"

	. "github.com/pulsar-beam/src/icrypto"
)

func TestAES(t *testing.T) {
	//https://golang.org/src/crypto/aes/cipher.go
	Key := "test1234test1234" //only 16/24/32 key length is supported
	e := AES{DefaultSalt: Key}
	var s string
	s = "this is a test string"
	text := []byte(s)
	key := []byte(Key)

	ciphertext, err := e.Encrypt(text, key)
	errNil(t, err)

	decryptedText, err := e.Decrypt(ciphertext, key)
	errNil(t, err)
	equals(t, text, decryptedText)
}

func TestRSA(t *testing.T) {
	//https://golang.org/src/crypto/rsa/rsa_test.go
	rsa, err := NewRSA()
	errNil(t, err)
	var s string
	s = "this is a test string"
	text := []byte(s)

	ciphertext, err := rsa.EncryptWithDefaultKey(text)
	errNil(t, err)

	decryptedText, err := rsa.DecryptWithDefaultKey(ciphertext)
	errNil(t, err)
	equals(t, text, decryptedText)

	bytes, err := rsa.GetPrivateKey()
	pstr := string(bytes)
	errNil(t, err)
	assert(t, len(pstr) > 512, "min length")

	//test to be encrypted by another object
	_, err = rsa.GetPrivateKey()
	errNil(t, err)

	pubKey, err := rsa.GetPublicKey()
	errNil(t, err)

	rsaDe, err := NewRSAWithKeys(nil, pubKey)
	errNil(t, err)

	text2 := []byte("another test to encrypted by one obj then decrypted by the original obj")
	ciphertext2, err := rsaDe.EncryptWithDefaultKey(text2)
	errNil(t, err)

	decrypted2, err := rsa.DecryptWithDefaultKey(ciphertext2)
	errNil(t, err)
	equals(t, text2, decrypted2)
}

func TestController64EncodeWithEncryption(t *testing.T) {
	s := "mockpassword1234"
	en, err := EncryptWithBase64(s)
	errNil(t, err)

	de, err2 := DecryptWithBase64(en)
	errNil(t, err2)
	equals(t, s, de)
}

func TestGenWriteKey(t *testing.T) {
	// Create 1 million to make sure no duplicates
	size := 1000000
	set := make(map[string]bool)
	for i := 0; i < size; i++ {
		id := GenTopicKey()
		set[id] = true
	}
	assert(t, size == len(set), "WriteKey duplicates found")

}

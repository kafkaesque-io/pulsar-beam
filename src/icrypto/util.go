package icrypto

// encryption and decryption utility functions
import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

var e AES

const defaultSymKey string = "popl4190LKOI4862" //16 character length

var defaultRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// init
func init() {
	e = AES{DefaultSalt: defaultSymKey}
}

// EncryptWithBase64 encrypts a string with AES default key and returns 64encoded string
func EncryptWithBase64(str string) (string, error) {
	text := []byte(str)
	encrypted, err := e.EncryptWithDefaultKey(text)
	if err != nil {
		return "", err
	}
	//fmt.Printf("original string %s \n", str)
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return encoded, nil
}

// DecryptWithBase64 a 64encoded string with the default key AES
func DecryptWithBase64(str string) (string, error) {
	decoded, err1 := base64.StdEncoding.DecodeString(str)
	if err1 != nil {
		fmt.Println("base64 decode error:", err1)
		return "", err1
	}
	decrypted, err := e.DecryptWithDefaultKey(decoded)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// RandKey generates a random key in n length
func RandKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = defaultRunes[rand.Intn(len(defaultRunes))]
	}
	return string(b)
}

// GenTopicKey generates a random key in 24 char length.
func GenTopicKey() string {
	return RandKey(24)
}

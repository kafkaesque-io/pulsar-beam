package tests

import (
	"fmt"
	"testing"
	"time"

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

	// test unsupported functions
	_, err = rsa.Decrypt([]byte{}, []byte{})
	equals(t, err.Error(), "unsupported")

	_, err = rsa.Encrypt([]byte{}, []byte{})
	equals(t, err.Error(), "unsupported")

}

func TestController64EncodeWithEncryption(t *testing.T) {
	s := "mockpassword1234"
	en, err := EncryptWithBase64(s)
	errNil(t, err)

	de, err2 := DecryptWithBase64(en)
	errNil(t, err2)
	equals(t, s, de)

	// failure cases
	_, err2 = DecryptWithBase64(fmt.Sprintf("%sy", en))
	assert(t, err2 != nil, "expect error with incorrect decryption")
}

func TestGenWriteKey(t *testing.T) {
	// Create 1 million to make sure no duplicates
	size := 100000
	set := make(map[string]bool)
	for i := 0; i < size; i++ {
		id := GenTopicKey()
		set[id] = true
	}
	assert(t, size == len(set), "WriteKey duplicates found")

}

func TestJWTRSASignAndVerify(t *testing.T) {
	// PK12 binary format
	testTokenSignAndVerify(t, "./pk12-binary-private.key", "./pk12-binary-public.key")

	// PEM format
	testTokenSignAndVerify(t, "./example_private_key", "./example_public_key.pub")
}

func testTokenSignAndVerify(t *testing.T, privateKeyPath, publicKeyPath string) {
	authen := NewRSAKeyPair(privateKeyPath, publicKeyPath)

	tokenString, err := authen.GenerateToken("myadmin")
	errNil(t, err)
	assert(t, len(tokenString) > 1, "a token string can be generated")

	token, err0 := authen.DecodeToken(tokenString)
	errNil(t, err0)
	assert(t, token.Valid, "validate a valid token")

	valid, _ := authen.VerifyTokenSubject("bogustokenstr", "myadmin")
	assert(t, valid == false, "validate token fails test")

	valid, _ = authen.VerifyTokenSubject(tokenString, "myadmin")
	assert(t, valid, "validate token's expected subject")

	valid, _ = authen.VerifyTokenSubject(tokenString, "admin")
	assert(t, valid == false, "validate token's mismatched subject")

	pulsarGeneratedToken := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA" // pragma: allowlist secret
	valid, _ = authen.VerifyTokenSubject(pulsarGeneratedToken, "picasso")
	assert(t, valid, "validate pulsar generated token and subject")

	subjects, err := authen.GetTokenSubject(pulsarGeneratedToken)
	errNil(t, err)
	equals(t, subjects, "picasso")

	t2 := time.Now().Add(time.Hour * 1)
	expireOffset := authen.GetTokenRemainingValidity(t2)
	equals(t, expireOffset, 3600)

}

package crypto

import (
	"fmt"
	"io"
	"log"
	"math/big"
	mr "math/rand"
	"strconv"
	"strings"
	"time"

	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/scrypt"
)

const (
	PW_SALT_BYTES = 32
	PW_HASH_BYTES = 64

	nonce = "ISWpnmdwRegMeNHXZNRegMISWpnmdwEPCPaafoNxfp"
)

var (
	key []byte
)

func SetKey(theKey []byte) {
	key = theKey
}

func Encrypt(theText string) string {
	// key := []byte("example key 1234")
	// key := []byte("Wroy7lzjLaXnAlND")
	for len(theText)%aes.BlockSize != 0 {
		theText += " "
	}
	// plaintext := []byte("exampleplaintext")
	plaintext := []byte(theText)

	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	if len(plaintext)%aes.BlockSize != 0 {
		log.Println("Error Encrypt: plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error Encrypt: aes", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Println("Error Encrypt: ReadFull", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.

	// return ciphertext
	return hex.EncodeToString(ciphertext)
}

func Decrypt(theBytes string) string {
	if len(theBytes) == 0 {
		return ""
	}

	ciphertext, _ := hex.DecodeString(theBytes)
	// ciphertext := theBytes

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error Decrypt: aes", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		log.Println("Error Decrypt: ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		log.Println("Error Decrypt: ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.

	retStr := string(ciphertext)
	return strings.TrimRight(retStr, " ")
	// return hex.DecodeString(ciphertext)
}

func hasStr(theText string, theSalt []byte) (hash []byte, err error) {
	hash, err = scrypt.Key([]byte(theText), theSalt, 1<<14, 8, 1, PW_HASH_BYTES)
	return
}

func MakeNonce(formUid int, userName string) string {
	secs := time.Now().Unix()
	randomString := RandomLetters(16) + ","

	if len(userName) > 0 {
		randomString += userName
	}

	randomString += "," + strconv.Itoa(formUid)
	randomString += "," + strconv.FormatInt(secs, 10) + ",3600"

	theLen := len(randomString)

	randomString += "," + strconv.Itoa(theLen)

	key := []byte(nonce)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(randomString))
	nonceString := fmt.Sprintf("%x", h.Sum(nil))

	return nonceString + "|" + randomString
}

func GetHash(theText, theSalt string) ([]byte, error) {
	var (
		hash []byte
		err  error
	)

	hash, err = scrypt.Key([]byte(theText), []byte(theSalt), 1<<14, 8, 1, PW_HASH_BYTES)
	if err != nil {
		return hash, err
	}
	return hash, nil
}

func RandomLetters(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	total := len(letters)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[randInt(0, total)]
	}
	return string(b)
}

func RandomChars(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	total := len(letters)
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[randInt(0, total)]
	}
	return string(b)
}

func randInt(min int, max int) int {
	var (
		n   *big.Int
		err error
	)
	diff := max - min
	if diff < 1 {
		diff = 1
	}
	maxVal := *big.NewInt(int64(diff))
	n, err = rand.Int(rand.Reader, &maxVal)
	if err != nil {
		mr.Seed(time.Now().Unix())
		val := mr.Intn(max - min)
		return val + min
	}

	return min + int(n.Int64())
}

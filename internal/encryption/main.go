package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func Encrypt(key string, data []byte) ([]byte, error) {

	var cipherData []byte

	aes, err := aes.NewCipher([]byte(key))
	if err != nil {
		return cipherData, errors.New(err.Error())
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return cipherData, errors.New(err.Error())
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return cipherData, errors.New(err.Error())
	}

	cipherData = gcm.Seal(nonce, nonce, data, nil)

	return cipherData, nil
}

func Decrypt(key string, data []byte) ([]byte, error) {

	aes, err := aes.NewCipher([]byte(key))
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return []byte{}, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	decryptedBytes, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return []byte{}, err
	}

	return decryptedBytes, nil

}

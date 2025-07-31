// pkg/security/crypto.go
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

// Chave de 32 bytes para AES-256. EM PRODUÇÃO IREI MUDAR PARA UM VAULT/KMS!
var aesKey = []byte("my-super-secret-32-byte-aes-key!")

// Chave secreta para HMAC. EM PRODUÇÃO IREI MUDAR PARA VAULT/KMS!
var hmacSecret = []byte("another-strong-secret-for-hmac!!")

// EncryptAES criptografa dados usando AES-256-GCM.
func EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptAES descriptografa dados usando AES-256-GCM.
func DecryptAES(encryptedString string) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// SignData gera uma assinatura HMAC-SHA256 para os dados.
func SignData(data []byte) string {
	h := hmac.New(sha256.New, hmacSecret)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifica se a assinatura HMAC-SHA256 é válida.
func VerifySignature(data []byte, signature string) bool {
	expectedSignature := SignData(data)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

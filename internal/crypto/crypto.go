package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/zoomxml/config"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrInvalidKeySize    = errors.New("invalid key size")
)

// getEncryptionKey returns the encryption key from config
func getEncryptionKey() []byte {
	cfg := config.Get()
	key := cfg.Auth.JWTSecret

	// Ensure key is exactly 32 bytes for AES-256
	if len(key) < 32 {
		// Pad with zeros if too short
		padded := make([]byte, 32)
		copy(padded, []byte(key))
		return padded
	}

	// Truncate if too long
	return []byte(key)[:32]
}

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key := getEncryptionKey()

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptCredentialData encrypts credential data based on type
func EncryptCredentialData(credType, login, password, token string) (string, error) {
	var data string

	switch credType {
	case "prefeitura_user_pass":
		if login == "" || password == "" {
			return "", errors.New("login and password are required for user/pass type")
		}
		data = fmt.Sprintf("%s:%s", login, password)
	case "prefeitura_token":
		if token == "" {
			return "", errors.New("token is required for token type")
		}
		data = token
	case "prefeitura_mixed":
		// Mixed type supports both user/pass and token
		if (login == "" || password == "") && token == "" {
			return "", errors.New("either login+password or token is required for mixed type")
		}
		if login != "" && password != "" && token != "" {
			// Both user/pass and token provided
			data = fmt.Sprintf("%s:%s:%s", login, password, token)
		} else if login != "" && password != "" {
			// Only user/pass provided
			data = fmt.Sprintf("%s:%s:", login, password)
		} else {
			// Only token provided
			data = fmt.Sprintf("::%s", token)
		}
	default:
		return "", fmt.Errorf("unsupported credential type: %s", credType)
	}

	return Encrypt(data)
}

// DecryptCredentialData decrypts credential data and returns login, password, token
func DecryptCredentialData(credType, encryptedData string) (login, password, token string, err error) {
	if encryptedData == "" {
		return "", "", "", nil
	}

	data, err := Decrypt(encryptedData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	switch credType {
	case "prefeitura_user_pass":
		// Data format: "login:password"
		parts := splitCredentialData(data, ":")
		if len(parts) != 2 {
			return "", "", "", errors.New("invalid user/pass credential format")
		}
		return parts[0], parts[1], "", nil
	case "prefeitura_token":
		// Data format: "token"
		return "", "", data, nil
	case "prefeitura_mixed":
		// Data format: "login:password:token" or "login:password:" or "::token"
		parts := splitCredentialData(data, ":")
		if len(parts) != 3 {
			return "", "", "", errors.New("invalid mixed credential format")
		}

		login = parts[0]
		password = parts[1]
		token = parts[2]

		// Clean up empty values
		if login == "" {
			login = ""
		}
		if password == "" {
			password = ""
		}
		if token == "" {
			token = ""
		}

		return login, password, token, nil
	default:
		return "", "", "", fmt.Errorf("unsupported credential type: %s", credType)
	}
}

// splitCredentialData safely splits credential data
func splitCredentialData(data, separator string) []string {
	if data == "" {
		return []string{}
	}

	// Simple split for now, could be enhanced for more complex formats
	result := []string{}
	current := ""

	for i, char := range data {
		if string(char) == separator {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}

		// Add the last part
		if i == len(data)-1 {
			result = append(result, current)
		}
	}

	return result
}

// Copyright 2025-2026 handlebargh and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package encryption provides AES-256-GCM encryption and decryption utilities
// for securing task and project data at rest.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

const (
	// KeySize is the size in bytes for AES-256 keys.
	KeySize = 32
	// NonceSize is the size in bytes for GCM nonces.
	NonceSize = 12
	// TagSize is the size in bytes for GCM authentication tags.
	TagSize = 16
	// EncryptedFilePrefix is prepended to encrypted file content for identification.
	EncryptedFilePrefix = "YATTO_ENC"
)

var (
	// ErrInvalidKey is returned when the encryption key is not 32 bytes.
	ErrInvalidKey = errors.New("encryption key must be 32 bytes for AES-256")
	// ErrDecryptionFailed is returned when decryption fails (wrong key or corrupted data).
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or key")
	// ErrNotEncrypted is returned when trying to decrypt data that is not encrypted.
	ErrNotEncrypted = errors.New("data is not encrypted")
)

// Encrypt encrypts plaintext data using AES-256-GCM with the provided key.
// Returns the encrypted ciphertext (nonce + ciphertext + tag) or an error.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Prepend our prefix for identification
	result := make([]byte, len(EncryptedFilePrefix)+len(ciphertext))
	copy(result, EncryptedFilePrefix)
	copy(result[len(EncryptedFilePrefix):], ciphertext)

	return result, nil
}

// Decrypt decrypts ciphertext data using AES-256-GCM with the provided key.
// The ciphertext should have been produced by Encrypt() and includes the nonce.
// Returns the decrypted plaintext or an error.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}

	// Check for encryption prefix
	if len(ciphertext) < len(EncryptedFilePrefix) ||
		string(ciphertext[:len(EncryptedFilePrefix)]) != EncryptedFilePrefix {
		return nil, ErrNotEncrypted
	}

	actualCiphertext := ciphertext[len(EncryptedFilePrefix):]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(actualCiphertext) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, ciphertextBytes := actualCiphertext[:nonceSize], actualCiphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// IsEncrypted checks if the given data has the encryption prefix.
func IsEncrypted(data []byte) bool {
	return len(data) >= len(EncryptedFilePrefix) &&
		string(data[:len(EncryptedFilePrefix)]) == EncryptedFilePrefix
}

// GenerateKey generates a new random 32-byte key suitable for AES-256.
// The key is returned as a base64-encoded string for easy storage.
func GenerateKey() (string, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// KeyFromBase64 decodes a base64-encoded key string into a byte slice.
// Returns an error if the key is not valid base64 or not 32 bytes.
func KeyFromBase64(base64Key string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	return key, nil
}

// EncryptFile encrypts a file's contents in place.
// Reads from the given path, encrypts the data, and writes it back.
func EncryptFile(path string, key []byte) error {
	data, err := osReadFile(path)
	if err != nil {
		return err
	}

	// Skip if already encrypted
	if IsEncrypted(data) {
		return nil
	}

	encrypted, err := Encrypt(data, key)
	if err != nil {
		return err
	}

	return osWriteFile(path, encrypted, 0o600)
}

// DecryptFile decrypts a file's contents in place.
// Reads from the given path, decrypts the data, and writes it back.
func DecryptFile(path string, key []byte) error {
	data, err := osReadFile(path)
	if err != nil {
		return err
	}

	// Skip if not encrypted
	if !IsEncrypted(data) {
		return nil
	}

	decrypted, err := Decrypt(data, key)
	if err != nil {
		return err
	}

	return osWriteFile(path, decrypted, 0o600)
}

// osReadFile is a wrapper for os.ReadFile to allow mocking in tests
var osReadFile = os.ReadFile

// osWriteFile is a wrapper for os.WriteFile to allow mocking in tests
var osWriteFile = os.WriteFile

// EncryptToBase64 encrypts data and returns the result as a base64 string.
// Useful for embedding encrypted data in structured formats.
func EncryptToBase64(plaintext, key []byte) (string, error) {
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptFromBase64 decrypts a base64-encoded ciphertext string.
func DecryptFromBase64(base64Ciphertext string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		return nil, err
	}
	return Decrypt(ciphertext, key)
}

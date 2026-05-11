package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"get-license-be/internal/httperr"
)

type Crypto struct {
	publicKey ed25519.PublicKey
	bundleKey []byte
}

func NewCrypto(publicKeyB64 string, bundleKeyB64 string) (*Crypto, error) {
	publicKey, err := loadPublicKey(publicKeyB64)
	if err != nil {
		return nil, err
	}
	bundleKey, err := loadBundleKey(bundleKeyB64)
	if err != nil {
		return nil, err
	}
	return &Crypto{publicKey: publicKey, bundleKey: bundleKey}, nil
}

func (c *Crypto) DecryptBundle(fileContent string) (Bundle, error) {
	var out Bundle
	rawEnvelope, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return out, httperr.New(400, "Invalid file content. Could not decode base64 or parse JSON.")
	}
	var envelope BundleEnvelope
	if err := json.Unmarshal(rawEnvelope, &envelope); err != nil {
		return out, httperr.New(400, "Invalid file content. Could not decode base64 or parse JSON.")
	}
	if envelope.V != 1 || envelope.Alg != "AES-GCM-256" {
		return out, httperr.New(400, "Unsupported bundle format or algorithm.")
	}
	iv, err1 := base64.StdEncoding.DecodeString(envelope.IV)
	tag, err2 := base64.StdEncoding.DecodeString(envelope.Tag)
	ciphertext, err3 := base64.StdEncoding.DecodeString(envelope.Data)
	if err1 != nil || err2 != nil || err3 != nil {
		return out, httperr.New(400, "Invalid file content. Could not decode base64 or parse JSON.")
	}
	block, err := aes.NewCipher(c.bundleKey)
	if err != nil {
		return out, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return out, err
	}
	sealed := append(ciphertext, tag...)
	plain, err := gcm.Open(nil, iv, sealed, nil)
	if err != nil {
		return out, httperr.New(400, "Decryption failed. BUNDLE_KEY is likely incorrect or data is corrupted.")
	}
	if err := json.Unmarshal(plain, &out); err != nil {
		return out, httperr.New(400, "Decryption failed. BUNDLE_KEY is likely incorrect or data is corrupted.")
	}
	return out, nil
}

func (c *Crypto) VerifyToken(token string) (map[string]any, error) {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	sig, err := hex.DecodeString(envelope.Signature)
	if err != nil {
		return nil, err
	}
	if !ed25519.Verify(c.publicKey, []byte(envelope.Data), sig) {
		return nil, fmt.Errorf("INVALID_SIGNATURE")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(envelope.Data), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func loadPublicKey(raw string) (ed25519.PublicKey, error) {
	if raw == "" {
		return nil, fmt.Errorf("PUBLIC_KEY is not configured in the environment")
	}
	keyDER, err := base64.StdEncoding.DecodeString(stripWhitespace(raw))
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(keyDER)
	if err != nil {
		return nil, err
	}
	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("PUBLIC_KEY must be an Ed25519 SPKI public key")
	}
	return edKey, nil
}

func loadBundleKey(raw string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(key) != 32 {
		return nil, httperr.New(500, "BUNDLE_KEY is not configured in the environment.")
	}
	return key, nil
}

func stripWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, `\n`, "")), "")
}

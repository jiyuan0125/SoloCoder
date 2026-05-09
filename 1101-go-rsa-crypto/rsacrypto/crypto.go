package rsacrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"log"

	"rsa-crypto/common"
)

var (
	ErrPlaintextTooLong       = errors.New("plaintext too long for the key size and padding mode")
	ErrUnsupportedPadding     = errors.New("unsupported padding mode")
	ErrUnsupportedHashAlgo    = errors.New("unsupported hash algorithm")
	ErrCiphertextInvalid      = errors.New("invalid ciphertext")
	ErrDecryptionFailed       = errors.New("decryption failed")
	ErrEncryptionFailed       = errors.New("encryption failed")
	ErrPublicKeyRequired      = errors.New("public key is required")
	ErrPrivateKeyRequired     = errors.New("private key is required")
)

func getHash(hashAlgo common.HashAlgorithm) (hash.Hash, crypto.Hash, error) {
	switch hashAlgo {
	case common.HashSHA1:
		return sha1.New(), crypto.SHA1, nil
	case common.HashSHA224:
		h := sha256.New224()
		return h, crypto.SHA224, nil
	case common.HashSHA256, "":
		return sha256.New(), crypto.SHA256, nil
	case common.HashSHA384:
		return sha512.New384(), crypto.SHA384, nil
	case common.HashSHA512:
		return sha512.New(), crypto.SHA512, nil
	default:
		return nil, crypto.Hash(0), fmt.Errorf("%w: %s", ErrUnsupportedHashAlgo, hashAlgo)
	}
}

func Encrypt(pub *rsa.PublicKey, plaintext []byte, padding common.PaddingMode, hashAlgo common.HashAlgorithm, label string) (string, error) {
	if pub == nil {
		return "", ErrPublicKeyRequired
	}

	keySize := pub.Size() * 8
	var maxSize int
	var ciphertext []byte
	var err error

	switch padding {
	case common.PaddingPKCS1v15:
		maxSize = MaxPlaintextSize(keySize, "pkcs1v15")
		if len(plaintext) > maxSize {
			return "", fmt.Errorf("%w: plaintext length is %d bytes, maximum allowed for PKCS#1 v1.5 with %d-bit key is %d bytes",
				ErrPlaintextTooLong, len(plaintext), keySize, maxSize)
		}
		ciphertext, err = rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
		}

	case common.PaddingOAEP:
		hash, _, err := getHash(hashAlgo)
		if err != nil {
			return "", err
		}
		hashSize := hash.Size()
		maxSize = keySize/8 - 2*hashSize - 2
		if len(plaintext) > maxSize {
			return "", fmt.Errorf("%w: plaintext length is %d bytes, maximum allowed for OAEP with %d-bit key and %s hash is %d bytes",
				ErrPlaintextTooLong, len(plaintext), keySize, hashAlgo, maxSize)
		}
		if label != "" {
			log.Printf("Encrypting with OAEP, label length: %d", len(label))
		}
		ciphertext, err = rsa.EncryptOAEP(hash, rand.Reader, pub, plaintext, []byte(label))
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
		}

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedPadding, padding)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(priv *rsa.PrivateKey, ciphertextB64 string, padding common.PaddingMode, hashAlgo common.HashAlgorithm, label string) ([]byte, error) {
	if priv == nil {
		return nil, ErrPrivateKeyRequired
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCiphertextInvalid, err)
	}

	switch padding {
	case common.PaddingPKCS1v15:
		plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
		}
		return plaintext, nil

	case common.PaddingOAEP:
		hash, _, err := getHash(hashAlgo)
		if err != nil {
			return nil, err
		}
		if label != "" {
			log.Printf("Decrypting with OAEP, label length: %d", len(label))
		}
		plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, []byte(label))
		if err != nil {
			if label != "" {
				log.Printf("Decryption failed - possible label mismatch. Used label: %s", label)
			}
			return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
		}
		return plaintext, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPadding, padding)
	}
}

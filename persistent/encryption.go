package persistent

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type encryption struct {
	base BlockStorage
	aead cipher.AEAD
}

// WithEncryption wraps a BlockStorage implementation and makes sure that all
// values are encrypted with AES-GCM before being processed further.
//
// The encryption key is derived with PBKDF2 from `password`.
func WithEncryption(base BlockStorage, password string) (BlockStorage, error) {
	key := pbkdf2.Key([]byte(password), []byte("7fedd6d671beec56"), 4096, 32, sha1.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &encryption{base, aead}, nil
}

func (e *encryption) Start(ctx context.Context) error { return e.base.Start(ctx) }

func (e *encryption) Get(ctx context.Context, ptr uint64) ([]byte, error) {
	raw, err := e.base.Get(ctx, ptr)
	if err != nil {
		return nil, err
	}
	ns := e.aead.NonceSize()
	if len(raw) < ns {
		return nil, fmt.Errorf("storage: ciphertext is too small")
	}
	return e.aead.Open(nil, raw[:ns], raw[ns:], []byte(fmt.Sprintf("%x", ptr)))
}

func (e *encryption) Set(ctx context.Context, ptr uint64, data []byte) error {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	ct := e.aead.Seal(nil, nonce, data, []byte(fmt.Sprintf("%x", ptr)))

	return e.base.Set(ctx, ptr, append(nonce, ct...))
}

func (e *encryption) Commit(ctx context.Context) error { return e.base.Commit(ctx) }
func (e *encryption) Rollback(ctx context.Context)     { e.base.Rollback(ctx) }

package pkg

import (
	"context"
	"crypto/rand"
	"crypto/rsa"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
)

type KeyPair interface {
	PublicJWK() (map[string]interface{}, error)
}

type keyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

func (kp *keyPair) PublicJWK() (map[string]interface{}, error) {
	key, err := jwk.New(kp.PublicKey)
	if err != nil {
		return nil, err
	}

	key.Set(jwk.KeyIDKey, kp.KeyID)

	return key.AsMap(context.Background())
}

var KeyPairHolder KeyPair

func GenerateKeyPair() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	publicKey := &privateKey.PublicKey

	KeyPairHolder = &keyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KeyID:      uuid.New().String(),
	}

	return nil
}

package pkg

import (
	"context"
	"crypto/rand"
	"crypto/rsa"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
)

func (app *Application) UseServerSecurity() *Application {
	kp, _ := generateKeyPair()
	app.KeyPair = kp
	return app
}

type KeyPair interface {
	PublicJWK() (map[string]interface{}, error)
	GetPrivateKey() *rsa.PrivateKey
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

func (kp *keyPair) GetPrivateKey() *rsa.PrivateKey {
	return kp.PrivateKey
}

func generateKeyPair() (KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	publicKey := &privateKey.PublicKey

	return &keyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KeyID:      uuid.New().String(),
	}, nil
}

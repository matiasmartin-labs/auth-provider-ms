package jwks

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// KeyPair holds an RSA keypair with an associated key ID.
type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	KeyID      string
}

// NewKeyPair generates a key ID and wraps the provided private key.
func NewKeyPair(privateKey *rsa.PrivateKey) *KeyPair {
	return &KeyPair{
		PrivateKey: privateKey,
		KeyID:      uuid.NewString(),
	}
}

// NewJwksHandler returns a Gin handler that serves the RSA public key as a JWKS JSON response.
//
// Deprecated: prefer NewJwksHandlerFromPublicKey, which accepts the public key and
// key ID directly so that KeyPair is no longer required.
func NewJwksHandler(kp *KeyPair) gin.HandlerFunc {
	return NewJwksHandlerFromPublicKey(&kp.PrivateKey.PublicKey, kp.KeyID)
}

// NewJwksHandlerFromPublicKey returns a Gin handler that serves pub as a JWKS JSON response
// identified by keyID. Use this variant together with app.GetRSAPublicKey() and app.GetRSAKeyID()
// so that the local KeyPair wrapper is not needed.
func NewJwksHandlerFromPublicKey(pub *rsa.PublicKey, keyID string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"keys": []gin.H{
				{
					"kty": "RSA",
					"kid": keyID,
					"use": "sig",
					"alg": "RS256",
					"n":   base64URLEncode(pub.N),
					"e":   base64URLEncodeInt(pub.E),
				},
			},
		})
	}
}

func base64URLEncode(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}

func base64URLEncodeInt(e int) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(e)) //nolint:gosec
	// trim leading zero bytes
	for len(b) > 1 && b[0] == 0 {
		b = b[1:]
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

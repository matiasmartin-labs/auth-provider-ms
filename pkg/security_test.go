package pkg

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)
	assert.NotNil(t, kp)
}

func TestGenerateKeyPair_ReturnsValidPrivateKey(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)
	privateKey := kp.GetPrivateKey()
	assert.NotNil(t, privateKey)
	assert.IsType(t, &rsa.PrivateKey{}, privateKey)
}

func TestGenerateKeyPair_GeneratesUniqueKeyIDs(t *testing.T) {
	kp1, err1 := generateKeyPair()
	kp2, err2 := generateKeyPair()

	require.NoError(t, err1)
	require.NoError(t, err2)

	internal1 := kp1.(*keyPair)
	internal2 := kp2.(*keyPair)

	assert.NotEqual(t, internal1.KeyID, internal2.KeyID)
}

func TestGenerateKeyPair_GeneratesUniqueKeys(t *testing.T) {
	kp1, err1 := generateKeyPair()
	kp2, err2 := generateKeyPair()

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, kp1.GetPrivateKey(), kp2.GetPrivateKey())
}

func TestKeyPair_GetPrivateKey(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)
	privateKey := kp.GetPrivateKey()

	assert.NotNil(t, privateKey)
	assert.NotNil(t, privateKey.PublicKey)
	assert.Equal(t, 2048, privateKey.N.BitLen())
}

func TestKeyPair_PublicJWK(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)
	jwk, err := kp.PublicJWK()

	require.NoError(t, err)
	assert.NotNil(t, jwk)

	assert.Contains(t, jwk, "kty")
	assert.Contains(t, jwk["kty"].(interface{ String() string }).String(), "RSA")
	assert.Contains(t, jwk, "kid")
	assert.Contains(t, jwk, "n")
	assert.Contains(t, jwk, "e")
}

func TestKeyPair_PublicJWK_HasKeyID(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)
	jwk, err := kp.PublicJWK()

	require.NoError(t, err)

	kid, ok := jwk["kid"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, kid)

	internal := kp.(*keyPair)
	assert.Equal(t, internal.KeyID, kid)
}

func TestKeyPair_PublicJWK_Consistency(t *testing.T) {
	kp, err := generateKeyPair()

	require.NoError(t, err)

	jwk1, err1 := kp.PublicJWK()
	jwk2, err2 := kp.PublicJWK()

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.Equal(t, jwk1["kid"], jwk2["kid"])
	assert.Equal(t, jwk1["n"], jwk2["n"])
	assert.Equal(t, jwk1["e"], jwk2["e"])
}

func TestApplication_UseServerSecurity(t *testing.T) {
	app := &Application{}

	result := app.UseServerSecurity()

	assert.Same(t, app, result) // Returns same instance for chaining
	assert.NotNil(t, app.KeyPair)
}

func TestApplication_UseServerSecurity_KeyPairIsUsable(t *testing.T) {
	app := &Application{}
	app.UseServerSecurity()

	privateKey := app.KeyPair.GetPrivateKey()
	assert.NotNil(t, privateKey)

	jwk, err := app.KeyPair.PublicJWK()
	assert.NoError(t, err)
	assert.NotNil(t, jwk)
}

func TestApplication_UseServerSecurity_ChainedCalls(t *testing.T) {
	app := &Application{}

	result := app.UseServerSecurity()
	assert.NotNil(t, result.KeyPair)
}

func BenchmarkGenerateKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateKeyPair()
	}
}

func BenchmarkPublicJWK(b *testing.B) {
	kp, _ := generateKeyPair()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kp.PublicJWK()
	}
}

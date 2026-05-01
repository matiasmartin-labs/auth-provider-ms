package me

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	fwkerrors "github.com/matiasmartin-labs/common-fwk/errors"
	fwkclaims "github.com/matiasmartin-labs/common-fwk/security/claims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMeHandler_Success(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		cl := fwkclaims.Claims{
			Email:   "test@example.com",
			Name:    "Test User",
			Picture: "https://example.com/picture.jpg",
			Roles:   []string{"USER"},
			Subject: "test@example.com",
		}
		c.Set("claims", cl)
		c.Next()
	}, MeHandler)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response MeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "https://example.com/picture.jpg", response.Picture)
}

func TestMeHandler_NoClaims(t *testing.T) {
	router := gin.New()
	router.GET("/me", MeHandler)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var payload map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	require.NoError(t, err)

	assert.Equal(t, fwkerrors.CodeClaimsMissing, payload["code"])
	assert.Equal(t, "no authentication claims found", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestMeHandler_InvalidClaimsFormat(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		// Store a non-claims value to trigger the type assertion failure path.
		// GetClaims returns (Claims{}, false) when the type doesn't match,
		// which is the same as missing claims — both return 401 CodeClaimsMissing.
		c.Set("claims", "invalid-claims-format")
		c.Next()
	}, MeHandler)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// GetClaims returns false for wrong type — treated same as missing claims.
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var payload map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	require.NoError(t, err)

	assert.Equal(t, fwkerrors.CodeClaimsMissing, payload["code"])
	assert.Equal(t, "no authentication claims found", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestMeHandler_EmptyFields(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		cl := fwkclaims.Claims{
			Email:   "",
			Name:    "",
			Picture: "",
		}
		c.Set("claims", cl)
		c.Next()
	}, MeHandler)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response MeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Empty(t, response.Email)
	assert.Empty(t, response.Name)
	assert.Empty(t, response.Picture)
}

func TestMeResponse_JSONStructure(t *testing.T) {
	response := MeResponse{
		Email:   "user@example.com",
		Name:    "John Doe",
		Picture: "https://example.com/john.jpg",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 3)
	assert.Contains(t, parsed, "email")
	assert.Contains(t, parsed, "name")
	assert.Contains(t, parsed, "picture")
}

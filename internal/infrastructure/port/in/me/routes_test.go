package me

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMeHandler_Success(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		claims := &pkg.Claims{
			Email:   "test@example.com",
			Name:    "Test User",
			Picture: "https://example.com/picture.jpg",
			Roles:   []string{"USER"},
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "test@example.com",
			},
		}
		c.Set("claims", claims)
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
	router.GET("/me", MeHandler) // Sin middleware que agregue claims

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var payload map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	require.NoError(t, err)

	assert.Equal(t, pkg.AuthCodeClaimsMissing, payload["code"])
	assert.Equal(t, "no authentication claims found", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestMeHandler_InvalidClaimsFormat(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("claims", "invalid-claims-format")
		c.Next()
	}, MeHandler)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var payload map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	require.NoError(t, err)

	assert.Equal(t, pkg.AuthCodeClaimsInvalid, payload["code"])
	assert.Equal(t, "invalid authentication claims", payload["message"])
	assert.Len(t, payload, 2)
	_, hasLegacyError := payload["error"]
	assert.False(t, hasLegacyError)
}

func TestMeHandler_EmptyFields(t *testing.T) {
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		claims := &pkg.Claims{
			Email:   "",
			Name:    "",
			Picture: "",
		}
		c.Set("claims", claims)
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

package jwt_test

import (
	"crypto/ecdsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtFlow(t *testing.T) {
	// Generate a key pair
	privateKey, publicKey, err := jwt.GenerateKeyPair()
	require.NoError(t, err)

	// Generate access and refresh tokens
	username := "testuser"
	id := 1

	accessToken, err := jwt.GetAccessToken(username, privateKey, id)
	require.NoError(t, err)

	refreshToken, err := jwt.GetRefreshToken(username, privateKey, id)
	require.NoError(t, err)

	// Generate JWKS JSON
	jwksJSON, err := jwt.GenerateJwksJSON(publicKey)
	require.NoError(t, err)

	// Serve the JWKS JSON
	server := httptest.NewServer(jwt.JWKSHandler(jwksJSON))
	defer server.Close()

	// Consume the JWKS JSON
	response, err := http.Get(server.URL + "/" + jwt.JWKSPath)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	var jwks jwt.JWKS
	err = json.NewDecoder(response.Body).Decode(&jwks)
	require.NoError(t, err)
	require.Len(t, jwks.Keys, 1)

	// Extract the public key from the JWKS JSON
	consumedPublicKey, err := jwt.PublicKeyFromJwksJSON(jwksJSON)
	require.NoError(t, err)

	// Validate the access token
	valid, claims, err := jwt.Validate(accessToken, func() (*ecdsa.PublicKey, error) { return consumedPublicKey, nil })
	require.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, id, claims.ID)

	// Validate the refresh token
	valid, claims, err = jwt.Validate(refreshToken, func() (*ecdsa.PublicKey, error) { return consumedPublicKey, nil })
	require.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, id, claims.ID)
}

func TestInvalidToken(t *testing.T) {
	// Generate a key pair
	privateKey, publicKey, err := jwt.GenerateKeyPair()
	require.NoError(t, err)

	// Generate access token
	username := "testuser"
	id := 1

	accessToken, err := jwt.GetAccessToken(username, privateKey, id)
	require.NoError(t, err)

	// Modify the token to make it invalid
	invalidToken := accessToken[:len(accessToken)-5] + "abcde"

	// Validate the invalid token
	valid, claims, err := jwt.Validate(invalidToken, func() (*ecdsa.PublicKey, error) { return publicKey, nil })
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, claims)
}

func TestJWTMiddlewareWithValidToken(t *testing.T) {
	// Generate a key pair
	privateKey, publicKey, err := jwt.GenerateKeyPair()
	require.NoError(t, err)

	// Generate access token
	username := "testuser"
	id := 1

	accessToken, err := jwt.GetAccessToken(username, privateKey, id)
	require.NoError(t, err)

	// Mock publicKeyFunc to return a public key.
	mockPublicKeyFunc := func() (*ecdsa.PublicKey, error) {
		// You can use a pre-generated public key here or generate a new key pair for testing.
		return publicKey, nil
	}

	// Create a Gin test context and a response recorder.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(jwt.JWT(mockPublicKeyFunc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, model.Result{
			Success: common_err.SUCCESS,
			Message: "success",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", accessToken)
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Assert the response status code and content.
	assert.Equal(t, http.StatusOK, respRecorder.Code)

	result := model.Result{}
	err = json.Unmarshal(respRecorder.Body.Bytes(), &result)

	assert.Equal(t, result.Success, common_err.SUCCESS)
	require.NoError(t, err)
}

func TestJWTMiddlewareWithInvalidToken(t *testing.T) {
	// Generate a key pair
	_, publicKey, err := jwt.GenerateKeyPair()
	require.NoError(t, err)

	// Mock publicKeyFunc to return a public key.
	mockPublicKeyFunc := func() (*ecdsa.PublicKey, error) {
		// You can use a pre-generated public key here or generate a new key pair for testing.
		return publicKey, nil
	}

	// Create a Gin test context and a response recorder.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(jwt.JWT(mockPublicKeyFunc))
	router.Use(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.GET("/test", func(c *gin.Context) {
		assert.Fail(t, "this handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid_token")
	respRecorder := httptest.NewRecorder()

	router.ServeHTTP(respRecorder, req)

	// Assert the response status code and content.
	assert.Equal(t, http.StatusUnauthorized, respRecorder.Code)

	result := model.Result{}
	err = json.Unmarshal(respRecorder.Body.Bytes(), &result)

	assert.Equal(t, result.Success, common_err.ERROR_AUTH_TOKEN)
	require.NoError(t, err)
}

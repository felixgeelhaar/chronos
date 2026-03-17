package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ascend/api/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthIntegration_RegistrationFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	t.Run("successful user registration", func(t *testing.T) {
		// Prepare registration request
		regReq := dto.RegisterRequest{
			Email:    "john@example.com",
			Password: "SecurePassword123!",
			Name:     "John Doe",
		}

		body, err := json.Marshal(regReq)
		require.NoError(t, err)

		// Make registration request
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var authResp dto.AuthResponse
		err = json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)

		assert.NotEmpty(t, authResp.User.ID)
		assert.Equal(t, "john@example.com", authResp.User.Email)
		assert.Equal(t, "John Doe", authResp.User.Name)
		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.NotZero(t, authResp.ExpiresIn)
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		// Register first user
		regReq := dto.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "Password123!",
			Name:     "First User",
		}

		body, _ := json.Marshal(regReq)
		resp, _ := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		resp.Body.Close()

		// Try to register with same email
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return conflict error
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("weak password validation", func(t *testing.T) {
		regReq := dto.RegisterRequest{
			Email:    "weak@example.com",
			Password: "weak",
			Name:     "Test User",
		}

		body, _ := json.Marshal(regReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return bad request for weak password
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid email format", func(t *testing.T) {
		regReq := dto.RegisterRequest{
			Email:    "not-an-email",
			Password: "SecurePassword123!",
			Name:     "Test User",
		}

		body, _ := json.Marshal(regReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuthIntegration_LoginFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register a test user first
	regReq := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "TestPassword123!",
		Name:     "Test User",
	}
	body, _ := json.Marshal(regReq)
	resp, _ := http.Post(
		ts.TestServer.URL+"/v1/auth/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	resp.Body.Close()

	t.Run("successful login with correct credentials", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}

		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp dto.AuthResponse
		err = json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)

		assert.Equal(t, "test@example.com", authResp.User.Email)
		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
	})

	t.Run("login with incorrect password", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "WrongPassword123!",
		}

		body, _ := json.Marshal(loginReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("login with non-existent email", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "SomePassword123!",
		}

		body, _ := json.Marshal(loginReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthIntegration_TokenRefreshFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register and login to get tokens
	regReq := dto.RegisterRequest{
		Email:    "refresh@example.com",
		Password: "RefreshTest123!",
		Name:     "Refresh User",
	}
	body, _ := json.Marshal(regReq)
	resp, _ := http.Post(
		ts.TestServer.URL+"/v1/auth/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	defer resp.Body.Close()

	var initialAuth dto.AuthResponse
	json.NewDecoder(resp.Body).Decode(&initialAuth)

	t.Run("successful token refresh", func(t *testing.T) {
		refreshReq := dto.RefreshTokenRequest{
			RefreshToken: initialAuth.RefreshToken,
		}

		body, err := json.Marshal(refreshReq)
		require.NoError(t, err)

		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/refresh",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp dto.AuthResponse
		err = json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)

		// Verify we got valid tokens back
		assert.NotEmpty(t, authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)

		// Verify the refreshed token works by accessing a protected endpoint
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+authResp.AccessToken)

		client := &http.Client{}
		meResp, err := client.Do(req)
		require.NoError(t, err)
		defer meResp.Body.Close()

		assert.Equal(t, http.StatusOK, meResp.StatusCode, "Refreshed token should work for protected endpoints")
	})

	t.Run("token refresh with invalid refresh token", func(t *testing.T) {
		refreshReq := dto.RefreshTokenRequest{
			RefreshToken: "invalid-token-here",
		}

		body, _ := json.Marshal(refreshReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/refresh",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthIntegration_ProtectedEndpointAccess(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register a user to get access token
	regReq := dto.RegisterRequest{
		Email:    "protected@example.com",
		Password: "ProtectedTest123!",
		Name:     "Protected User",
	}
	body, _ := json.Marshal(regReq)
	resp, _ := http.Post(
		ts.TestServer.URL+"/v1/auth/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	defer resp.Body.Close()

	var authResp dto.AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)

	t.Run("access protected endpoint with valid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+authResp.AccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var userResp dto.UserResponse
		err = json.NewDecoder(resp.Body).Decode(&userResp)
		require.NoError(t, err)

		assert.Equal(t, "protected@example.com", userResp.Email)
		assert.Equal(t, "Protected User", userResp.Name)
	})

	t.Run("access protected endpoint without token", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("access protected endpoint with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer invalid-token-here")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("access protected endpoint with malformed authorization header", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "InvalidFormat "+authResp.AccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthIntegration_CompleteUserJourney(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	t.Run("complete user journey: register, login, refresh, access protected", func(t *testing.T) {
		// Step 1: Register
		regReq := dto.RegisterRequest{
			Email:    "journey@example.com",
			Password: "JourneyTest123!",
			Name:     "Journey User",
		}
		body, _ := json.Marshal(regReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/auth/register",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var registerResp dto.AuthResponse
		json.NewDecoder(resp.Body).Decode(&registerResp)
		resp.Body.Close()

		initialRefreshToken := registerResp.RefreshToken

		// Step 2: Login
		loginReq := dto.LoginRequest{
			Email:    "journey@example.com",
			Password: "JourneyTest123!",
		}
		body, _ = json.Marshal(loginReq)
		resp, err = http.Post(
			ts.TestServer.URL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp dto.AuthResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)
		resp.Body.Close()

		// Step 3: Access protected endpoint
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)

		client := &http.Client{}
		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// Step 4: Refresh token
		refreshReq := dto.RefreshTokenRequest{
			RefreshToken: initialRefreshToken,
		}
		body, _ = json.Marshal(refreshReq)
		resp, err = http.Post(
			ts.TestServer.URL+"/v1/auth/refresh",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResp dto.AuthResponse
		json.NewDecoder(resp.Body).Decode(&refreshResp)
		resp.Body.Close()

		// Verify we got valid tokens
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)

		// Step 5: Access protected endpoint with new token
		req, _ = http.NewRequest("GET", ts.TestServer.URL+"/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)

		resp, err = client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var userResp dto.UserResponse
		json.NewDecoder(resp.Body).Decode(&userResp)
		resp.Body.Close()

		assert.Equal(t, "journey@example.com", userResp.Email)
		assert.Equal(t, "Journey User", userResp.Name)
	})
}

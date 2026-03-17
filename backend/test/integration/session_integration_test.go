package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionIntegration_CreateSession(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register a user and get access token
	token := registerAndGetToken(t, ts, "session@example.com", "SessionTest123!", "Session User")

	t.Run("successful session creation with sets", func(t *testing.T) {
		notes := "Great workout today"
		rpe := 8.5

		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now(),
			Notes: &notes,
			Sets: []dto.CreateSetInput{
				{
					ExerciseName: "Bench Press",
					Weight:       100.0,
					Reps:         10,
					RPE:          &rpe,
					SetOrder:     1,
				},
				{
					ExerciseName: "Bench Press",
					Weight:       100.0,
					Reps:         8,
					RPE:          &rpe,
					SetOrder:     2,
				},
			},
		}

		body, err := json.Marshal(sessionReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var sessionResp dto.SessionResponse
		err = json.NewDecoder(resp.Body).Decode(&sessionResp)
		require.NoError(t, err)

		assert.NotEmpty(t, sessionResp.ID)
		assert.NotEmpty(t, sessionResp.UserID)
		assert.Equal(t, "Great workout today", *sessionResp.Notes)
		assert.Len(t, sessionResp.Sets, 2)
		assert.Equal(t, "Bench Press", sessionResp.Sets[0].ExerciseName)
		assert.Equal(t, 100.0, sessionResp.Sets[0].Weight)
		assert.Equal(t, 10, sessionResp.Sets[0].Reps)
		assert.Equal(t, 8.5, *sessionResp.Sets[0].RPE)
	})

	t.Run("create session without sets", func(t *testing.T) {
		sessionReq := dto.CreateSessionRequest{
			Date: time.Now(),
		}

		body, _ := json.Marshal(sessionReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var sessionResp dto.SessionResponse
		json.NewDecoder(resp.Body).Decode(&sessionResp)

		assert.NotEmpty(t, sessionResp.ID)
		assert.Len(t, sessionResp.Sets, 0)
	})

	t.Run("create session without authentication", func(t *testing.T) {
		sessionReq := dto.CreateSessionRequest{
			Date: time.Now(),
		}

		body, _ := json.Marshal(sessionReq)
		resp, err := http.Post(
			ts.TestServer.URL+"/v1/sessions",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("create session with invalid data", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"sets": []map[string]interface{}{
				{
					"exercise_name": "Bench Press",
					"weight":        -50.0, // Invalid: negative weight
					"reps":          10,
					"set_order":     1,
				},
			},
		}

		body, _ := json.Marshal(invalidReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestSessionIntegration_GetSession(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "getsession@example.com", "GetTest123!", "Get User")

	// Create a session
	sessionID := createTestSession(t, ts, token, time.Now())

	t.Run("get existing session", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sessionResp dto.SessionResponse
		err = json.NewDecoder(resp.Body).Decode(&sessionResp)
		require.NoError(t, err)

		assert.Equal(t, sessionID, sessionResp.ID)
		assert.NotEmpty(t, sessionResp.UserID)
	})

	t.Run("get non-existent session", func(t *testing.T) {
		fakeID := uuid.New()
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+fakeID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get session with invalid ID format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/invalid-uuid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get session without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestSessionIntegration_ListSessions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "listsessions@example.com", "ListTest123!", "List User")

	// Create multiple sessions
	now := time.Now()
	sessionID1 := createTestSession(t, ts, token, now.Add(-48*time.Hour))
	sessionID2 := createTestSession(t, ts, token, now.Add(-24*time.Hour))
	sessionID3 := createTestSession(t, ts, token, now)

	t.Run("list all sessions", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp dto.SessionListResponse
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.Len(t, listResp.Sessions, 3)
		assert.Equal(t, 3, listResp.Total)
		assert.Equal(t, 1, listResp.Page)
		assert.Equal(t, 20, listResp.PageSize)
		assert.Equal(t, 1, listResp.TotalPages)

		// Verify sessions are returned (order may vary)
		sessionIDs := []uuid.UUID{sessionID1, sessionID2, sessionID3}
		for _, session := range listResp.Sessions {
			assert.Contains(t, sessionIDs, session.ID)
		}
	})

	t.Run("list sessions with pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions?page=1&page_size=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp dto.SessionListResponse
		json.NewDecoder(resp.Body).Decode(&listResp)

		assert.Len(t, listResp.Sessions, 2)
		assert.Equal(t, 3, listResp.Total)
		assert.Equal(t, 1, listResp.Page)
		assert.Equal(t, 2, listResp.PageSize)
		assert.Equal(t, 2, listResp.TotalPages)
	})

	t.Run("list sessions page 2", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions?page=2&page_size=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp dto.SessionListResponse
		json.NewDecoder(resp.Body).Decode(&listResp)

		assert.Len(t, listResp.Sessions, 1)
		assert.Equal(t, 3, listResp.Total)
		assert.Equal(t, 2, listResp.Page)
	})

	t.Run("list sessions without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestSessionIntegration_UpdateSession(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "updatesession@example.com", "UpdateTest123!", "Update User")

	// Create a session
	sessionID := createTestSession(t, ts, token, time.Now())

	t.Run("update session date and notes", func(t *testing.T) {
		newDate := time.Now().Add(24 * time.Hour)
		newNotes := "Updated workout notes"

		updateReq := dto.UpdateSessionRequest{
			Date:  &newDate,
			Notes: &newNotes,
		}

		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sessionResp dto.SessionResponse
		json.NewDecoder(resp.Body).Decode(&sessionResp)

		assert.Equal(t, sessionID, sessionResp.ID)
		assert.Equal(t, "Updated workout notes", *sessionResp.Notes)
	})

	t.Run("update session with new sets", func(t *testing.T) {
		rpe := 9.0
		updateReq := dto.UpdateSessionRequest{
			Sets: []dto.CreateSetInput{
				{
					ExerciseName: "Squat",
					Weight:       150.0,
					Reps:         5,
					RPE:          &rpe,
					SetOrder:     1,
				},
			},
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var sessionResp dto.SessionResponse
		json.NewDecoder(resp.Body).Decode(&sessionResp)

		assert.Len(t, sessionResp.Sets, 1)
		assert.Equal(t, "Squat", sessionResp.Sets[0].ExerciseName)
		assert.Equal(t, 150.0, sessionResp.Sets[0].Weight)
	})

	t.Run("update non-existent session", func(t *testing.T) {
		fakeID := uuid.New()
		newNotes := "Should fail"

		updateReq := dto.UpdateSessionRequest{
			Notes: &newNotes,
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+fakeID.String(), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("update session without authentication", func(t *testing.T) {
		newNotes := "Unauthorized update"

		updateReq := dto.UpdateSessionRequest{
			Notes: &newNotes,
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestSessionIntegration_DeleteSession(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "deletesession@example.com", "DeleteTest123!", "Delete User")

	t.Run("delete existing session", func(t *testing.T) {
		// Create a session to delete
		sessionID := createTestSession(t, ts, token, time.Now())

		req, err := http.NewRequest("DELETE", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify session is deleted
		getReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		getReq.Header.Set("Authorization", "Bearer "+token)

		getResp, _ := client.Do(getReq)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("delete non-existent session", func(t *testing.T) {
		fakeID := uuid.New()

		req, _ := http.NewRequest("DELETE", ts.TestServer.URL+"/v1/sessions/"+fakeID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete session without authentication", func(t *testing.T) {
		sessionID := createTestSession(t, ts, token, time.Now())

		req, _ := http.NewRequest("DELETE", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestSessionIntegration_Authorization(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register two users
	token1 := registerAndGetToken(t, ts, "user1@example.com", "User1Test123!", "User One")
	token2 := registerAndGetToken(t, ts, "user2@example.com", "User2Test123!", "User Two")

	// User 1 creates a session
	sessionID := createTestSession(t, ts, token1, time.Now())

	t.Run("user cannot get another user's session", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("user cannot update another user's session", func(t *testing.T) {
		newNotes := "Trying to update someone else's session"

		updateReq := dto.UpdateSessionRequest{
			Notes: &newNotes,
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token2)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("user cannot delete another user's session", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("user only sees their own sessions in list", func(t *testing.T) {
		// User 2 creates their own session
		createTestSession(t, ts, token2, time.Now())

		// User 2 lists sessions
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var listResp dto.SessionListResponse
		json.NewDecoder(resp.Body).Decode(&listResp)

		// User 2 should only see 1 session (their own)
		assert.Equal(t, 1, listResp.Total)
		assert.Len(t, listResp.Sessions, 1)
		assert.NotEqual(t, sessionID, listResp.Sessions[0].ID)
	})
}

func TestSessionIntegration_CompleteSessionJourney(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user
	token := registerAndGetToken(t, ts, "journey@example.com", "JourneyTest123!", "Journey User")

	t.Run("complete session lifecycle", func(t *testing.T) {
		// Step 1: Create session with sets
		rpe := 7.5
		notes := "First workout"

		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now(),
			Notes: &notes,
			Sets: []dto.CreateSetInput{
				{
					ExerciseName: "Deadlift",
					Weight:       200.0,
					Reps:         5,
					RPE:          &rpe,
					SetOrder:     1,
				},
			},
		}

		body, _ := json.Marshal(sessionReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResp dto.SessionResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		resp.Body.Close()

		sessionID := createResp.ID

		// Step 2: Get the session
		getReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		getReq.Header.Set("Authorization", "Bearer "+token)

		getResp, _ := client.Do(getReq)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var getSessionResp dto.SessionResponse
		json.NewDecoder(getResp.Body).Decode(&getSessionResp)
		getResp.Body.Close()

		assert.Equal(t, sessionID, getSessionResp.ID)

		// Step 3: Update the session
		updatedNotes := "Updated workout notes"
		updateReq := dto.UpdateSessionRequest{
			Notes: &updatedNotes,
		}

		updateBody, _ := json.Marshal(updateReq)
		putReq, _ := http.NewRequest("PUT", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), bytes.NewBuffer(updateBody))
		putReq.Header.Set("Authorization", "Bearer "+token)
		putReq.Header.Set("Content-Type", "application/json")

		updateResp, _ := client.Do(putReq)
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedSessionResp dto.SessionResponse
		json.NewDecoder(updateResp.Body).Decode(&updatedSessionResp)
		updateResp.Body.Close()

		assert.Equal(t, "Updated workout notes", *updatedSessionResp.Notes)

		// Step 4: List sessions
		listReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions", nil)
		listReq.Header.Set("Authorization", "Bearer "+token)

		listResp, _ := client.Do(listReq)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var listSessionResp dto.SessionListResponse
		json.NewDecoder(listResp.Body).Decode(&listSessionResp)
		listResp.Body.Close()

		assert.Equal(t, 1, listSessionResp.Total)

		// Step 5: Delete the session
		delReq, _ := http.NewRequest("DELETE", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		delReq.Header.Set("Authorization", "Bearer "+token)

		delResp, _ := client.Do(delReq)
		assert.Equal(t, http.StatusNoContent, delResp.StatusCode)
		delResp.Body.Close()

		// Step 6: Verify deletion
		verifyReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/sessions/"+sessionID.String(), nil)
		verifyReq.Header.Set("Authorization", "Bearer "+token)

		verifyResp, _ := client.Do(verifyReq)
		assert.Equal(t, http.StatusNotFound, verifyResp.StatusCode)
		verifyResp.Body.Close()
	})
}

// Helper functions

// registerAndGetToken is a helper that registers a user and returns their access token
func registerAndGetToken(t *testing.T, ts *TestServer, email, password, name string) string {
	regReq := dto.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	body, _ := json.Marshal(regReq)
	resp, err := http.Post(
		ts.TestServer.URL+"/v1/auth/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to register user")

	var authResp dto.AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)

	return authResp.AccessToken
}

// createTestSession is a helper that creates a session and returns its ID
func createTestSession(t *testing.T, ts *TestServer, token string, date time.Time) uuid.UUID {
	notes := "Test session"
	rpe := 7.0

	sessionReq := dto.CreateSessionRequest{
		Date:  date,
		Notes: &notes,
		Sets: []dto.CreateSetInput{
			{
				ExerciseName: "Test Exercise",
				Weight:       100.0,
				Reps:         10,
				RPE:          &rpe,
				SetOrder:     1,
			},
		},
	}

	body, _ := json.Marshal(sessionReq)
	req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, fmt.Sprintf("Failed to create test session: %v", err))

	var sessionResp dto.SessionResponse
	json.NewDecoder(resp.Body).Decode(&sessionResp)

	return sessionResp.ID
}

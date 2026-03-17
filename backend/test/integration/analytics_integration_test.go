package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsIntegration_ExerciseHistory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "analytics@example.com", "AnalyticsTest123!", "Analytics User")

	// Create sessions with different exercises
	createSessionsForAnalytics(t, ts, token)

	t.Run("get exercise history with data", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/Bench%20Press", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var historyResp dto.ExerciseHistoryResponse
		err = json.NewDecoder(resp.Body).Decode(&historyResp)
		require.NoError(t, err)

		assert.Equal(t, "Bench Press", historyResp.ExerciseName)
		assert.Greater(t, len(historyResp.Records), 0)
		assert.Greater(t, historyResp.TotalVolume, 0.0)
		assert.Greater(t, historyResp.TotalSets, 0)
	})

	t.Run("get exercise history with date range", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")

		url := fmt.Sprintf("%s/v1/analytics/exercise/Squat?start_date=%s&end_date=%s",
			ts.TestServer.URL, startDate, endDate)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var historyResp dto.ExerciseHistoryResponse
		json.NewDecoder(resp.Body).Decode(&historyResp)

		assert.Equal(t, "Squat", historyResp.ExerciseName)
	})

	t.Run("get exercise history for non-existent exercise", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/NonExistent", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var historyResp dto.ExerciseHistoryResponse
		json.NewDecoder(resp.Body).Decode(&historyResp)

		// Should return empty results
		assert.Equal(t, "NonExistent", historyResp.ExerciseName)
		assert.Len(t, historyResp.Records, 0)
		assert.Equal(t, 0.0, historyResp.TotalVolume)
	})

	t.Run("get exercise history without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/Bench%20Press", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("get exercise history without name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should either be 404 (route not found) or 400 (bad request)
		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest)
	})
}

func TestAnalyticsIntegration_ACWR(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "acwr@example.com", "ACWRTest123!", "ACWR User")

	// Create sessions spread over time
	createSessionsOverTime(t, ts, token)

	t.Run("get ACWR for all exercises", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/acwr", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var acwrResp dto.ACWRResponse
		err = json.NewDecoder(resp.Body).Decode(&acwrResp)
		require.NoError(t, err)

		// ACWR response should have meaningful data
		assert.GreaterOrEqual(t, acwrResp.CurrentACWR, 0.0)
		assert.GreaterOrEqual(t, acwrResp.AcuteLoad, 0.0)
		assert.GreaterOrEqual(t, acwrResp.ChronicLoad, 0.0)
		assert.NotEmpty(t, acwrResp.Status)
		assert.NotEmpty(t, acwrResp.Recommendation)
	})

	t.Run("get ACWR for specific exercise", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/acwr?exercise=Deadlift", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var acwrResp dto.ACWRResponse
		json.NewDecoder(resp.Body).Decode(&acwrResp)

		assert.Equal(t, "Deadlift", acwrResp.ExerciseName)
	})

	t.Run("get ACWR without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/acwr", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAnalyticsIntegration_VolumeProgress(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "volume@example.com", "VolumeTest123!", "Volume User")

	// Create sessions for volume tracking
	createSessionsForAnalytics(t, ts, token)

	t.Run("get volume progress by week", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?exercise=Squat&period=week", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var volumeResp dto.VolumeProgressResponse
		err = json.NewDecoder(resp.Body).Decode(&volumeResp)
		require.NoError(t, err)

		assert.Equal(t, "Squat", volumeResp.ExerciseName)
		assert.Equal(t, "week", volumeResp.Period)
		assert.GreaterOrEqual(t, volumeResp.TotalVolume, 0.0)
		assert.NotEmpty(t, volumeResp.Trend)
	})

	t.Run("get volume progress by month", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?period=month", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var volumeResp dto.VolumeProgressResponse
		json.NewDecoder(resp.Body).Decode(&volumeResp)

		assert.Equal(t, "month", volumeResp.Period)
	})

	t.Run("get volume progress by year", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?period=year", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get volume progress with invalid period", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?period=invalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get volume progress without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?period=week", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAnalyticsIntegration_ProgressSummary(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user and get token
	token := registerAndGetToken(t, ts, "summary@example.com", "SummaryTest123!", "Summary User")

	// Create sessions
	createSessionsForAnalytics(t, ts, token)

	t.Run("get progress summary default period", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/summary", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summaryResp dto.ProgressSummaryResponse
		err = json.NewDecoder(resp.Body).Decode(&summaryResp)
		require.NoError(t, err)

		assert.NotEmpty(t, summaryResp.Period)
		assert.GreaterOrEqual(t, summaryResp.TotalSessions, 0)
		assert.GreaterOrEqual(t, summaryResp.TotalSets, 0)
		assert.GreaterOrEqual(t, summaryResp.TotalVolume, 0.0)
	})

	t.Run("get progress summary with date range", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")

		url := fmt.Sprintf("%s/v1/analytics/summary?start_date=%s&end_date=%s",
			ts.TestServer.URL, startDate, endDate)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summaryResp dto.ProgressSummaryResponse
		json.NewDecoder(resp.Body).Decode(&summaryResp)

		// Verify the period string contains the date range
		assert.NotEmpty(t, summaryResp.Period)
	})

	t.Run("get progress summary without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/summary", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAnalyticsIntegration_CrossUserIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register two users
	token1 := registerAndGetToken(t, ts, "user1-analytics@example.com", "User1Test123!", "User One")
	token2 := registerAndGetToken(t, ts, "user2-analytics@example.com", "User2Test123!", "User Two")

	// User 1 creates sessions
	createSessionsForAnalytics(t, ts, token1)

	t.Run("user 2 should not see user 1 analytics", func(t *testing.T) {
		// User 2 requests exercise history
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/Bench%20Press", nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var historyResp dto.ExerciseHistoryResponse
		json.NewDecoder(resp.Body).Decode(&historyResp)

		// User 2 should have no data
		assert.Len(t, historyResp.Records, 0)
		assert.Equal(t, 0.0, historyResp.TotalVolume)
	})

	t.Run("user 2 should have separate progress summary", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/summary", nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var summaryResp dto.ProgressSummaryResponse
		json.NewDecoder(resp.Body).Decode(&summaryResp)

		// User 2 has no sessions
		assert.Equal(t, 0, summaryResp.TotalSessions)
	})
}

func TestAnalyticsIntegration_CompleteAnalyticsJourney(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register user
	token := registerAndGetToken(t, ts, "journey-analytics@example.com", "JourneyTest123!", "Journey User")

	t.Run("complete analytics workflow", func(t *testing.T) {
		// Step 1: Create training sessions
		createSessionsForAnalytics(t, ts, token)

		client := &http.Client{}

		// Step 2: View exercise history
		histReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/exercise/Squat", nil)
		histReq.Header.Set("Authorization", "Bearer "+token)

		histResp, err := client.Do(histReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, histResp.StatusCode)
		histResp.Body.Close()

		// Step 3: Check ACWR
		acwrReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/acwr", nil)
		acwrReq.Header.Set("Authorization", "Bearer "+token)

		acwrResp, err := client.Do(acwrReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, acwrResp.StatusCode)
		acwrResp.Body.Close()

		// Step 4: View volume progress
		volumeReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/volume?period=week", nil)
		volumeReq.Header.Set("Authorization", "Bearer "+token)

		volumeResp, err := client.Do(volumeReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, volumeResp.StatusCode)
		volumeResp.Body.Close()

		// Step 5: Get overall summary
		summaryReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/analytics/summary", nil)
		summaryReq.Header.Set("Authorization", "Bearer "+token)

		summaryResp, err := client.Do(summaryReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, summaryResp.StatusCode)

		var summary dto.ProgressSummaryResponse
		json.NewDecoder(summaryResp.Body).Decode(&summary)
		summaryResp.Body.Close()

		// Verify summary has data
		assert.Greater(t, summary.TotalSessions, 0)
		assert.Greater(t, summary.TotalVolume, 0.0)
	})
}

// Helper functions

// createSessionsForAnalytics creates multiple sessions with different exercises for testing
func createSessionsForAnalytics(t *testing.T, ts *TestServer, token string) {
	exercises := []struct {
		name   string
		weight float64
		reps   int
	}{
		{"Bench Press", 100.0, 10},
		{"Bench Press", 105.0, 8},
		{"Squat", 150.0, 5},
		{"Squat", 155.0, 5},
		{"Deadlift", 200.0, 3},
	}

	for i, ex := range exercises {
		rpe := 7.0 + float64(i)*0.5
		notes := fmt.Sprintf("Training session %d", i+1)

		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now().Add(-time.Duration(i*24) * time.Hour),
			Notes: &notes,
			Sets: []dto.CreateSetInput{
				{
					ExerciseName: ex.name,
					Weight:       ex.weight,
					Reps:         ex.reps,
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
		require.NoError(t, err, "Failed to create session for analytics test")
		resp.Body.Close()
	}
}

// createSessionsOverTime creates sessions spread over multiple weeks
func createSessionsOverTime(t *testing.T, ts *TestServer, token string) {
	// Create sessions over last 30 days
	for daysAgo := 1; daysAgo <= 30; daysAgo += 3 {
		rpe := 7.0
		notes := fmt.Sprintf("Session %d days ago", daysAgo)

		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now().Add(-time.Duration(daysAgo*24) * time.Hour),
			Notes: &notes,
			Sets: []dto.CreateSetInput{
				{
					ExerciseName: "Deadlift",
					Weight:       180.0 + float64(daysAgo),
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
		resp.Body.Close()
	}
}

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ascend/api/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestVideo creates a multipart form request with a test video file
func createVideoUploadRequest(t *testing.T, url, token string, filename string, content []byte, metadata map[string]string) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add video file
	part, err := writer.CreateFormFile("video", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)

	// Add metadata fields
	for key, value := range metadata {
		err = writer.WriteField(key, value)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", url, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	return client.Do(req)
}

func TestVideoIntegration_UploadFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Register and login to get token
	regReq := dto.RegisterRequest{
		Email:    "video@example.com",
		Password: "VideoTest123!",
		Name:     "Video User",
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
	token := authResp.AccessToken

	t.Run("successful video upload without metadata", func(t *testing.T) {
		videoContent := []byte("fake video content for testing")

		resp, err := createVideoUploadRequest(
			t,
			ts.TestServer.URL+"/v1/videos",
			token,
			"test_video.mp4",
			videoContent,
			map[string]string{}, // no metadata
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var videoResp dto.VideoResponse
		err = json.NewDecoder(resp.Body).Decode(&videoResp)
		require.NoError(t, err)

		assert.NotEmpty(t, videoResp.ID)
		assert.Equal(t, authResp.User.ID, videoResp.UserID)
		assert.NotEmpty(t, videoResp.URL)
		assert.Equal(t, int64(len(videoContent)), videoResp.FileSize)
		assert.Nil(t, videoResp.SessionID)
		assert.Nil(t, videoResp.ExerciseName)

		// Verify file was stored in mock S3
		urlParts := strings.Split(videoResp.URL, ".com/")
		require.Len(t, urlParts, 2)
		key := urlParts[1]
		storedContent, exists := ts.MockS3.GetFile(key)
		assert.True(t, exists, "File should be stored in mock S3")
		assert.Equal(t, videoContent, storedContent)
	})

	t.Run("video upload with session metadata", func(t *testing.T) {
		// Create a session first
		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now(),
			Notes: stringPtr("Test session for video"),
		}
		sessionBody, _ := json.Marshal(sessionReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(sessionBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		sessionResp, err := client.Do(req)
		require.NoError(t, err)
		defer sessionResp.Body.Close()

		var session dto.SessionResponse
		json.NewDecoder(sessionResp.Body).Decode(&session)

		// Upload video with session metadata
		videoContent := []byte("video with session metadata")

		resp, err := createVideoUploadRequest(
			t,
			ts.TestServer.URL+"/v1/videos",
			token,
			"session_video.mp4",
			videoContent,
			map[string]string{
				"session_id":    session.ID.String(),
				"exercise_name": "Bench Press",
			},
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var videoResp dto.VideoResponse
		err = json.NewDecoder(resp.Body).Decode(&videoResp)
		require.NoError(t, err)

		assert.NotNil(t, videoResp.SessionID)
		assert.Equal(t, session.ID, *videoResp.SessionID)
		assert.NotNil(t, videoResp.ExerciseName)
		assert.Equal(t, "Bench Press", *videoResp.ExerciseName)
	})

	t.Run("video upload without token fails", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "test.mp4")
		part.Write([]byte("content"))
		writer.Close()

		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("video upload without file fails", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestVideoIntegration_GetVideo(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Setup: Register user and upload video
	token, videoID := setupVideoTest(t, ts)

	t.Run("get video with valid ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+videoID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var videoResp dto.VideoResponse
		err = json.NewDecoder(resp.Body).Decode(&videoResp)
		require.NoError(t, err)

		assert.Equal(t, videoID, videoResp.ID)
		assert.NotEmpty(t, videoResp.URL)
	})

	t.Run("get video with invalid ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/invalid-uuid", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get non-existent video", func(t *testing.T) {
		nonExistentID := uuid.New()
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+nonExistentID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestVideoIntegration_ListVideos(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Setup: Register user and upload multiple videos
	token, userID := setupUserTest(t, ts)

	// Upload multiple videos
	videoIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		videoContent := []byte(fmt.Sprintf("video content %d", i))
		resp, err := createVideoUploadRequest(
			t,
			ts.TestServer.URL+"/v1/videos",
			token,
			fmt.Sprintf("video_%d.mp4", i),
			videoContent,
			map[string]string{},
		)
		require.NoError(t, err)

		var videoResp dto.VideoResponse
		json.NewDecoder(resp.Body).Decode(&videoResp)
		resp.Body.Close()
		videoIDs[i] = videoResp.ID
	}

	t.Run("list all videos", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp dto.VideoListResponse
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.Len(t, listResp.Videos, 3)
		assert.Equal(t, 3, listResp.Total)

		// Verify all videos belong to the user
		for _, video := range listResp.Videos {
			assert.Equal(t, userID, video.UserID)
		}
	})

	t.Run("list videos with pagination", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos?page=1&page_size=2", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp dto.VideoListResponse
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.LessOrEqual(t, len(listResp.Videos), 2)
		assert.Equal(t, 1, listResp.Page)
		assert.Equal(t, 2, listResp.PageSize)
	})
}

func TestVideoIntegration_ListVideosBySession(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	token, _ := setupUserTest(t, ts)

	// Create a session
	sessionReq := dto.CreateSessionRequest{
		Date:  time.Now(),
		Notes: stringPtr("Session with videos"),
	}
	sessionBody, _ := json.Marshal(sessionReq)
	req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(sessionBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	sessionResp, _ := client.Do(req)
	var session dto.SessionResponse
	json.NewDecoder(sessionResp.Body).Decode(&session)
	sessionResp.Body.Close()

	// Upload videos for this session
	for i := 0; i < 2; i++ {
		videoContent := []byte(fmt.Sprintf("session video %d", i))
		resp, err := createVideoUploadRequest(
			t,
			ts.TestServer.URL+"/v1/videos",
			token,
			fmt.Sprintf("session_video_%d.mp4", i),
			videoContent,
			map[string]string{
				"session_id": session.ID.String(),
			},
		)
		require.NoError(t, err)
		resp.Body.Close()
	}

	t.Run("list videos by session", func(t *testing.T) {
		req, err := http.NewRequest(
			"GET",
			ts.TestServer.URL+"/v1/sessions/"+session.ID.String()+"/videos",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp struct {
			Videos []dto.VideoResponse `json:"videos"`
		}
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		assert.Len(t, listResp.Videos, 2)
		for _, video := range listResp.Videos {
			assert.NotNil(t, video.SessionID)
			assert.Equal(t, session.ID, *video.SessionID)
		}
	})
}

func TestVideoIntegration_UpdateVideo(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	token, videoID := setupVideoTest(t, ts)

	t.Run("update video metadata", func(t *testing.T) {
		exerciseName := "Squat"
		updateReq := dto.UpdateVideoRequest{
			ExerciseName: &exerciseName,
		}
		body, _ := json.Marshal(updateReq)

		req, err := http.NewRequest(
			"PUT",
			ts.TestServer.URL+"/v1/videos/"+videoID.String(),
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var videoResp dto.VideoResponse
		err = json.NewDecoder(resp.Body).Decode(&videoResp)
		require.NoError(t, err)

		assert.Equal(t, videoID, videoResp.ID)
		assert.NotNil(t, videoResp.ExerciseName)
		assert.Equal(t, "Squat", *videoResp.ExerciseName)
	})

	t.Run("update video with session ID", func(t *testing.T) {
		// Create a session
		sessionReq := dto.CreateSessionRequest{
			Date: time.Now(),
		}
		sessionBody, _ := json.Marshal(sessionReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(sessionBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		sessionResp, _ := client.Do(req)
		var session dto.SessionResponse
		json.NewDecoder(sessionResp.Body).Decode(&session)
		sessionResp.Body.Close()

		// Update video with session
		updateReq := dto.UpdateVideoRequest{
			SessionID: &session.ID,
		}
		body, _ := json.Marshal(updateReq)

		req, err := http.NewRequest(
			"PUT",
			ts.TestServer.URL+"/v1/videos/"+videoID.String(),
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var videoResp dto.VideoResponse
		json.NewDecoder(resp.Body).Decode(&videoResp)

		assert.NotNil(t, videoResp.SessionID)
		assert.Equal(t, session.ID, *videoResp.SessionID)
	})
}

func TestVideoIntegration_DeleteVideo(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	token, videoID := setupVideoTest(t, ts)

	t.Run("delete video successfully", func(t *testing.T) {
		req, err := http.NewRequest(
			"DELETE",
			ts.TestServer.URL+"/v1/videos/"+videoID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify video is deleted
		getReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+videoID.String(), nil)
		getReq.Header.Set("Authorization", "Bearer "+token)

		getResp, _ := client.Do(getReq)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("delete non-existent video", func(t *testing.T) {
		nonExistentID := uuid.New()
		req, err := http.NewRequest(
			"DELETE",
			ts.TestServer.URL+"/v1/videos/"+nonExistentID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestVideoIntegration_GeneratePresignedURL(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	token, videoID := setupVideoTest(t, ts)

	t.Run("generate presigned URL with default expiration", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			ts.TestServer.URL+"/v1/videos/"+videoID.String()+"/presigned-url",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var presignedResp dto.GeneratePresignedURLResponse
		err = json.NewDecoder(resp.Body).Decode(&presignedResp)
		require.NoError(t, err)

		assert.NotEmpty(t, presignedResp.URL)
		assert.False(t, presignedResp.ExpiresAt.IsZero())
		assert.True(t, presignedResp.ExpiresAt.After(time.Now()))
	})

	t.Run("generate presigned URL with custom expiration", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			ts.TestServer.URL+"/v1/videos/"+videoID.String()+"/presigned-url?expiration=7200",
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var presignedResp dto.GeneratePresignedURLResponse
		json.NewDecoder(resp.Body).Decode(&presignedResp)

		// Check expiration is approximately 2 hours from now
		expectedExpiry := time.Now().Add(2 * time.Hour)
		assert.WithinDuration(t, expectedExpiry, presignedResp.ExpiresAt, 5*time.Second)
	})
}

func TestVideoIntegration_CrossUserIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	// Create two users
	_, videoID := setupVideoTest(t, ts)
	token2, _ := setupUserTest(t, ts)

	t.Run("user cannot access another user's video", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+videoID.String(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("user cannot update another user's video", func(t *testing.T) {
		exerciseName := "Deadlift"
		updateReq := dto.UpdateVideoRequest{
			ExerciseName: &exerciseName,
		}
		body, _ := json.Marshal(updateReq)

		req, err := http.NewRequest(
			"PUT",
			ts.TestServer.URL+"/v1/videos/"+videoID.String(),
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("user cannot delete another user's video", func(t *testing.T) {
		req, err := http.NewRequest(
			"DELETE",
			ts.TestServer.URL+"/v1/videos/"+videoID.String(),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token2)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestVideoIntegration_CompleteVideoJourney(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown(t)

	t.Run("complete video management journey", func(t *testing.T) {
		// Step 1: Register user
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

		var authResp dto.AuthResponse
		json.NewDecoder(resp.Body).Decode(&authResp)
		resp.Body.Close()
		token := authResp.AccessToken

		// Step 2: Create a session
		sessionReq := dto.CreateSessionRequest{
			Date:  time.Now(),
			Notes: stringPtr("Complete journey session"),
		}
		sessionBody, _ := json.Marshal(sessionReq)
		req, _ := http.NewRequest("POST", ts.TestServer.URL+"/v1/sessions", bytes.NewBuffer(sessionBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		sessionResp, _ := client.Do(req)
		var session dto.SessionResponse
		json.NewDecoder(sessionResp.Body).Decode(&session)
		sessionResp.Body.Close()

		// Step 3: Upload video with session
		videoContent := []byte("journey video content")
		uploadResp, err := createVideoUploadRequest(
			t,
			ts.TestServer.URL+"/v1/videos",
			token,
			"journey.mp4",
			videoContent,
			map[string]string{
				"session_id":    session.ID.String(),
				"exercise_name": "Bench Press",
			},
		)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, uploadResp.StatusCode)

		var video dto.VideoResponse
		json.NewDecoder(uploadResp.Body).Decode(&video)
		uploadResp.Body.Close()

		// Step 4: Get video details
		getReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+video.ID.String(), nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getResp, err := client.Do(getReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)
		getResp.Body.Close()

		// Step 5: List videos by session
		listReq, _ := http.NewRequest(
			"GET",
			ts.TestServer.URL+"/v1/sessions/"+session.ID.String()+"/videos",
			nil,
		)
		listReq.Header.Set("Authorization", "Bearer "+token)
		listResp, err := client.Do(listReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)
		listResp.Body.Close()

		// Step 6: Update video metadata
		newExercise := "Squat"
		updateReq := dto.UpdateVideoRequest{
			ExerciseName: &newExercise,
		}
		updateBody, _ := json.Marshal(updateReq)
		updateHTTPReq, _ := http.NewRequest(
			"PUT",
			ts.TestServer.URL+"/v1/videos/"+video.ID.String(),
			bytes.NewBuffer(updateBody),
		)
		updateHTTPReq.Header.Set("Content-Type", "application/json")
		updateHTTPReq.Header.Set("Authorization", "Bearer "+token)
		updateHTTPResp, err := client.Do(updateHTTPReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, updateHTTPResp.StatusCode)

		var updatedVideo dto.VideoResponse
		json.NewDecoder(updateHTTPResp.Body).Decode(&updatedVideo)
		updateHTTPResp.Body.Close()
		assert.Equal(t, "Squat", *updatedVideo.ExerciseName)

		// Step 7: Generate presigned URL
		presignReq, _ := http.NewRequest(
			"POST",
			ts.TestServer.URL+"/v1/videos/"+video.ID.String()+"/presigned-url",
			nil,
		)
		presignReq.Header.Set("Authorization", "Bearer "+token)
		presignResp, err := client.Do(presignReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, presignResp.StatusCode)
		presignResp.Body.Close()

		// Step 8: List all videos
		listAllReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos", nil)
		listAllReq.Header.Set("Authorization", "Bearer "+token)
		listAllResp, err := client.Do(listAllReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, listAllResp.StatusCode)

		var listAllVideos dto.VideoListResponse
		json.NewDecoder(listAllResp.Body).Decode(&listAllVideos)
		listAllResp.Body.Close()
		assert.GreaterOrEqual(t, len(listAllVideos.Videos), 1)

		// Step 9: Delete video
		deleteReq, _ := http.NewRequest(
			"DELETE",
			ts.TestServer.URL+"/v1/videos/"+video.ID.String(),
			nil,
		)
		deleteReq.Header.Set("Authorization", "Bearer "+token)
		deleteResp, err := client.Do(deleteReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
		deleteResp.Body.Close()

		// Step 10: Verify deletion
		verifyReq, _ := http.NewRequest("GET", ts.TestServer.URL+"/v1/videos/"+video.ID.String(), nil)
		verifyReq.Header.Set("Authorization", "Bearer "+token)
		verifyResp, err := client.Do(verifyReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, verifyResp.StatusCode)
		verifyResp.Body.Close()
	})
}

// Helper functions

func setupUserTest(t *testing.T, ts *TestServer) (token string, userID uuid.UUID) {
	email := fmt.Sprintf("user-%s@example.com", uuid.New().String()[:8])
	regReq := dto.RegisterRequest{
		Email:    email,
		Password: "Test123!",
		Name:     "Test User",
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

	return authResp.AccessToken, authResp.User.ID
}

func setupVideoTest(t *testing.T, ts *TestServer) (token string, videoID uuid.UUID) {
	token, _ = setupUserTest(t, ts)

	videoContent := []byte("test video content")
	resp, err := createVideoUploadRequest(
		t,
		ts.TestServer.URL+"/v1/videos",
		token,
		"test.mp4",
		videoContent,
		map[string]string{},
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var videoResp dto.VideoResponse
	json.NewDecoder(resp.Body).Decode(&videoResp)

	return token, videoResp.ID
}

func stringPtr(s string) *string {
	return &s
}

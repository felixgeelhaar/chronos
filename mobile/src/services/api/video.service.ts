import apiClient from './client';
import {
  VideoResponse,
  VideoUploadResponse,
  VideoListResponse,
  VideoAnalysisResponse,
} from './types';

/**
 * Video Service
 * Handles video upload, retrieval, and analysis operations
 */
class VideoService {
  /**
   * Upload a video file
   */
  async uploadVideo(
    file: {
      uri: string;
      type: string;
      name: string;
    },
    setId?: string,
    onProgress?: (progress: number) => void
  ): Promise<VideoUploadResponse> {
    const formData = new FormData();
    formData.append('video', {
      uri: file.uri,
      type: file.type,
      name: file.name,
    } as any);

    if (setId) {
      formData.append('set_id', setId);
    }

    return apiClient.upload<VideoUploadResponse>(
      '/v1/videos/upload',
      formData,
      onProgress
    );
  }

  /**
   * Get a specific video by ID
   */
  async getVideo(videoId: string): Promise<VideoResponse> {
    return apiClient.get<VideoResponse>(`/v1/videos/${videoId}`);
  }

  /**
   * List all videos for the current user
   */
  async listVideos(
    page: number = 1,
    pageSize: number = 20,
    setId?: string
  ): Promise<VideoListResponse> {
    const params: any = { page, page_size: pageSize };
    if (setId) params.set_id = setId;

    return apiClient.get<VideoListResponse>('/v1/videos', { params });
  }

  /**
   * Delete a video
   */
  async deleteVideo(videoId: string): Promise<void> {
    return apiClient.delete(`/v1/videos/${videoId}`);
  }

  /**
   * Get video analysis results
   */
  async getVideoAnalysis(videoId: string): Promise<VideoAnalysisResponse> {
    return apiClient.get<VideoAnalysisResponse>(`/v1/videos/${videoId}/analysis`);
  }

  /**
   * Trigger video reprocessing
   */
  async reprocessVideo(videoId: string): Promise<VideoResponse> {
    return apiClient.post<VideoResponse>(`/v1/videos/${videoId}/reprocess`, {});
  }

  /**
   * Get video playback URL
   */
  getVideoUrl(videoId: string, quality: 'original' | 'high' | 'medium' | 'low' = 'high'): string {
    const baseUrl = apiClient.getBaseURL();
    return `${baseUrl}/v1/videos/${videoId}/stream?quality=${quality}`;
  }

  /**
   * Get video thumbnail URL
   */
  getThumbnailUrl(videoId: string): string {
    const baseUrl = apiClient.getBaseURL();
    return `${baseUrl}/v1/videos/${videoId}/thumbnail`;
  }
}

// Export singleton instance
export const videoService = new VideoService();
export default videoService;

import apiClient from './client';
import {
  CreateSessionRequest,
  UpdateSessionRequest,
  SessionResponse,
  SessionListResponse,
} from './types';

/**
 * Session Service
 * Handles workout session operations
 */
class SessionService {
  /**
   * Create a new workout session
   */
  async createSession(data: CreateSessionRequest): Promise<SessionResponse> {
    return apiClient.post<SessionResponse>('/v1/sessions', data);
  }

  /**
   * Get a specific session by ID
   */
  async getSession(id: string): Promise<SessionResponse> {
    return apiClient.get<SessionResponse>(`/v1/sessions/${id}`);
  }

  /**
   * List all sessions for the current user
   */
  async listSessions(page: number = 1, pageSize: number = 20): Promise<SessionListResponse> {
    return apiClient.get<SessionListResponse>('/v1/sessions', {
      params: { page, page_size: pageSize },
    });
  }

  /**
   * Update an existing session
   */
  async updateSession(id: string, data: UpdateSessionRequest): Promise<SessionResponse> {
    return apiClient.put<SessionResponse>(`/v1/sessions/${id}`, data);
  }

  /**
   * Delete a session
   */
  async deleteSession(id: string): Promise<void> {
    return apiClient.delete(`/v1/sessions/${id}`);
  }
}

// Export singleton instance
export const sessionService = new SessionService();
export default sessionService;

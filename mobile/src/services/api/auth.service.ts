import apiClient from './client';
import {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  UserResponse,
} from './types';

/**
 * Authentication Service
 * Handles user authentication operations
 */
class AuthService {
  /**
   * Register a new user
   */
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/v1/auth/register', data);

    // Store tokens in the client
    await apiClient.setTokens(response.access_token, response.refresh_token);

    return response;
  }

  /**
   * Login with email and password
   */
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/v1/auth/login', data);

    // Store tokens in the client
    await apiClient.setTokens(response.access_token, response.refresh_token);

    return response;
  }

  /**
   * Logout the current user
   */
  async logout(): Promise<void> {
    await apiClient.clearTokens();
  }

  /**
   * Get current user profile
   */
  async getMe(): Promise<UserResponse> {
    return apiClient.get<UserResponse>('/v1/auth/me');
  }

  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    return apiClient.isAuthenticated();
  }
}

// Export singleton instance
export const authService = new AuthService();
export default authService;

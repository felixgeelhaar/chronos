import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080';

// Storage keys
const ACCESS_TOKEN_KEY = '@ascend:accessToken';
const REFRESH_TOKEN_KEY = '@ascend:refreshToken';

/**
 * API Client using Axios with automatic token refresh
 */
class APIClient {
  private client: AxiosInstance;
  private refreshing: boolean = false;
  private refreshSubscribers: ((token: string) => void)[] = [];

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  /**
   * Setup request and response interceptors
   */
  private setupInterceptors(): void {
    // Request interceptor: Add auth token to requests
    this.client.interceptors.request.use(
      async (config) => {
        const token = await AsyncStorage.getItem(ACCESS_TOKEN_KEY);
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor: Handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };

        // If error is 401 and we haven't retried yet
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.refreshing) {
            // If already refreshing, wait for the new token
            return new Promise((resolve) => {
              this.refreshSubscribers.push((token: string) => {
                if (originalRequest.headers) {
                  originalRequest.headers.Authorization = `Bearer ${token}`;
                }
                resolve(this.client(originalRequest));
              });
            });
          }

          originalRequest._retry = true;
          this.refreshing = true;

          try {
            const refreshToken = await AsyncStorage.getItem(REFRESH_TOKEN_KEY);
            if (!refreshToken) {
              throw new Error('No refresh token available');
            }

            // Attempt to refresh the token
            const response = await axios.post(`${API_BASE_URL}/v1/auth/refresh`, {
              refresh_token: refreshToken,
            });

            const { access_token, refresh_token: newRefreshToken } = response.data;

            // Store new tokens
            await AsyncStorage.setItem(ACCESS_TOKEN_KEY, access_token);
            await AsyncStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken);

            // Notify all waiting requests
            this.refreshSubscribers.forEach((callback) => callback(access_token));
            this.refreshSubscribers = [];

            // Retry the original request
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${access_token}`;
            }
            return this.client(originalRequest);
          } catch (refreshError) {
            // Refresh failed, clear tokens and redirect to login
            await this.clearTokens();
            return Promise.reject(refreshError);
          } finally {
            this.refreshing = false;
          }
        }

        return Promise.reject(error);
      }
    );
  }

  /**
   * Store authentication tokens
   */
  async setTokens(accessToken: string, refreshToken: string): Promise<void> {
    await AsyncStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    await AsyncStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }

  /**
   * Clear authentication tokens
   */
  async clearTokens(): Promise<void> {
    await AsyncStorage.removeItem(ACCESS_TOKEN_KEY);
    await AsyncStorage.removeItem(REFRESH_TOKEN_KEY);
  }

  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    const token = await AsyncStorage.getItem(ACCESS_TOKEN_KEY);
    return !!token;
  }

  /**
   * GET request
   */
  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  /**
   * POST request
   */
  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  /**
   * PUT request
   */
  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  /**
   * DELETE request
   */
  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  /**
   * Upload file (multipart/form-data)
   */
  async upload<T = any>(
    url: string,
    formData: FormData,
    onProgress?: (progress: number) => void,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const response = await this.client.post<T>(url, formData, {
      ...config,
      headers: {
        ...config?.headers,
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: onProgress
        ? (progressEvent) => {
            const percentCompleted = progressEvent.total
              ? Math.round((progressEvent.loaded * 100) / progressEvent.total)
              : 0;
            onProgress(percentCompleted);
          }
        : undefined,
    });
    return response.data;
  }

  /**
   * Get the base URL
   */
  getBaseURL(): string {
    return API_BASE_URL;
  }

  /**
   * Get the raw Axios instance (for advanced usage)
   */
  getInstance(): AxiosInstance {
    return this.client;
  }
}

// Export singleton instance
export const apiClient = new APIClient();
export default apiClient;

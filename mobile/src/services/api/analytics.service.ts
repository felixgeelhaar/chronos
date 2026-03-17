import apiClient from './client';
import {
  ExerciseHistoryResponse,
  ACWRResponse,
  VolumeProgressResponse,
  ProgressSummaryResponse,
} from './types';

/**
 * Analytics Service
 * Handles progress tracking and analytics operations
 */
class AnalyticsService {
  /**
   * Get exercise history with 1RM progression
   */
  async getExerciseHistory(
    exerciseName: string,
    startDate?: string,
    endDate?: string
  ): Promise<ExerciseHistoryResponse> {
    const params: any = {};
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;

    return apiClient.get<ExerciseHistoryResponse>(
      `/v1/analytics/exercise/${encodeURIComponent(exerciseName)}`,
      { params }
    );
  }

  /**
   * Get ACWR (Acute:Chronic Workload Ratio) for injury prevention
   */
  async getACWR(exerciseName?: string): Promise<ACWRResponse> {
    const params: any = {};
    if (exerciseName) params.exercise = exerciseName;

    return apiClient.get<ACWRResponse>('/v1/analytics/acwr', { params });
  }

  /**
   * Get volume progress over time
   */
  async getVolumeProgress(
    period: 'week' | 'month' | 'year' = 'month',
    exerciseName?: string
  ): Promise<VolumeProgressResponse> {
    const params: any = { period };
    if (exerciseName) params.exercise = exerciseName;

    return apiClient.get<VolumeProgressResponse>('/v1/analytics/volume', { params });
  }

  /**
   * Get overall progress summary
   */
  async getProgressSummary(startDate?: string, endDate?: string): Promise<ProgressSummaryResponse> {
    const params: any = {};
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;

    return apiClient.get<ProgressSummaryResponse>('/v1/analytics/summary', { params });
  }
}

// Export singleton instance
export const analyticsService = new AnalyticsService();
export default analyticsService;

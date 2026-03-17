/**
 * API Request and Response Types
 */

// Auth Types
export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  body_weight?: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: UserResponse;
}

export interface UserResponse {
  id: string;
  email: string;
  name: string;
  body_weight?: number;
  created_at: string;
  updated_at: string;
}

// Session Types
export interface CreateSessionRequest {
  date: string; // ISO 8601 date
  notes?: string;
  sets?: CreateSetInput[];
}

export interface CreateSetInput {
  exercise_name: string;
  weight: number;
  reps: number;
  rpe?: number;
  set_order: number;
  notes?: string;
}

export interface UpdateSessionRequest {
  date?: string;
  notes?: string;
}

export interface SessionResponse {
  id: string;
  user_id: string;
  date: string;
  notes?: string;
  sets: SetResponse[];
  total_volume: number;
  total_sets: number;
  created_at: string;
  updated_at: string;
}

export interface SetResponse {
  id: string;
  session_id: string;
  exercise_name: string;
  weight: number;
  reps: number;
  rpe?: number;
  set_order: number;
  notes?: string;
  volume: number;
  estimated_one_rep_max: number;
  created_at: string;
  updated_at: string;
}

export interface SessionListResponse {
  sessions: SessionResponse[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// One Rep Max Types
export interface CreateOneRepMaxRequest {
  exercise_name: string;
  weight: number;
  date: string;
}

export interface UpdateOneRepMaxRequest {
  weight?: number;
  date?: string;
}

export interface OneRepMaxResponse {
  id: string;
  user_id: string;
  exercise_name: string;
  weight: number;
  date: string;
  created_at: string;
  updated_at: string;
}

export interface OneRepMaxListResponse {
  records: OneRepMaxResponse[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface OneRepMaxHistoryResponse {
  exercise_name: string;
  current_record?: OneRepMaxResponse;
  history: OneRepMaxResponse[];
  personal_best: number;
  improvement: number;
}

// Analytics Types
export interface ExerciseHistoryResponse {
  exercise_name: string;
  start_date: string;
  end_date: string;
  records: ExerciseRecord[];
  personal_best_1rm: number;
  latest_1rm: number;
  improvement_percentage: number;
}

export interface ExerciseRecord {
  date: string;
  sets: SetResponse[];
  max_weight: number;
  total_volume: number;
  estimated_1rm: number;
  one_rep_max?: OneRepMaxRecord;
}

export interface OneRepMaxRecord {
  weight: number;
  date: string;
}

export interface ACWRResponse {
  exercise_name?: string;
  current_acwr: number;
  acute_load: number;
  chronic_load: number;
  status: 'optimal' | 'high_risk' | 'undertraining';
  recommendation: string;
  history?: ACWRDataPoint[];
}

export interface ACWRDataPoint {
  date: string;
  acwr: number;
  acute_load: number;
  chronic_load: number;
}

export interface VolumeProgressResponse {
  exercise_name?: string;
  period: 'week' | 'month' | 'year';
  data_points: VolumeDataPoint[];
  total_volume: number;
  average_volume: number;
  trend: 'increasing' | 'stable' | 'decreasing';
}

export interface VolumeDataPoint {
  period_start: string;
  period_end: string;
  volume: number;
  session_count: number;
}

export interface ProgressSummaryResponse {
  start_date: string;
  end_date: string;
  total_sessions: number;
  total_volume: number;
  average_session_volume: number;
  top_exercises: ExerciseVolumeSummary[];
  current_acwr: number;
  acwr_status: string;
}

export interface ExerciseVolumeSummary {
  exercise_name: string;
  total_volume: number;
  session_count: number;
  average_volume_per_session: number;
  max_weight: number;
}

// Video Types
export interface UploadVideoRequest {
  session_id?: string;
  exercise_name?: string;
}

export interface UpdateVideoRequest {
  session_id?: string;
  exercise_name?: string;
}

export interface VideoResponse {
  id: string;
  user_id: string;
  session_id?: string;
  url: string;
  thumbnail_url?: string;
  duration?: number;
  file_size: number;
  exercise_name?: string;
  date: string;
  created_at: string;
  updated_at: string;
}

export interface VideoListResponse {
  videos: VideoResponse[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface GeneratePresignedURLResponse {
  url: string;
  expires_at: string;
}

// Error Response
export interface APIError {
  error: {
    code: string;
    message: string;
    details?: string;
  };
}

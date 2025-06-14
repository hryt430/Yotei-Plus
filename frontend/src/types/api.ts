// === Common API Response Types ===
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data: T;
}

export interface ErrorResponse {
  error: string;
  code?: string;
}

// === Utility Types ===
export interface ApiError extends Error {
  status?: number;
  code?: string;
  response?: any;
}
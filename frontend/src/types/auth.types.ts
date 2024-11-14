export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
  name: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  message: string;
}

export interface RefreshTokenResponse {
  access_token: string;
}

export interface ApiError {
  message: string;
  error?: string;
}
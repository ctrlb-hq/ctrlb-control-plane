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
    token: string;
    message: string;
  }
  
  export interface ApiError {
    message: string;
    error?: string;
  }
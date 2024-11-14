import axios, { AxiosError } from 'axios';
import { LoginCredentials, RegisterCredentials, AuthResponse, RefreshTokenResponse, ApiError } from '../types/auth.types';

const apiUrl = import.meta.env.VITE_BACKEND_URI;
const API_BASE_URL = `${apiUrl}/api/auth/v1`;

const authService = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await axios.post(`${API_BASE_URL}/login`, credentials);
      const data = response.data;

      if (data.access_token && data.refresh_token) {
        localStorage.setItem('accessToken', data.access_token);
        localStorage.setItem('refreshToken', data.refresh_token);
      }

      return data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(axiosError.response?.data?.message || 'Login failed');
    }
  },

  register: async (userDetails: RegisterCredentials): Promise<AuthResponse> => {
    try {
      const response = await axios.post(`${API_BASE_URL}/register`, userDetails);
      const data = response.data;

      if (data.access_token && data.refresh_token) {
        localStorage.setItem('accessToken', data.access_token);
        localStorage.setItem('refreshToken', data.refresh_token);
      }

      return data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(axiosError.response?.data?.message || 'Registration failed');
    }
  },

  refreshToken: async (): Promise<RefreshTokenResponse | null> => {
    try {
      const refreshToken = localStorage.getItem('refreshToken');
      if (!refreshToken) {
        return null;
      }

      const response = await axios.post(`${API_BASE_URL}/refresh`, { refresh_token: refreshToken });
      const data = response.data;

      if (data.access_token) {
        localStorage.setItem('accessToken', data.access_token);
      }

      return data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      if (axiosError.response?.status === 401) {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
        return null;
      }
      throw new Error(axiosError.response?.data?.message || 'Token refresh failed');
    }
  },

  logout: async (): Promise<void> => {
    try {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
    } catch (error) {
      console.error('Logout error:', error);
      throw new Error('Logout failed');
    }
  },
};

export default authService;

import axios, { AxiosError } from 'axios';
import { LoginCredentials, RegisterCredentials, AuthResponse, ApiError } from '../types/auth.types';

const apiUrl = import.meta.env.VITE_BACKEND_URI;
const API_BASE_URL = `${apiUrl}/api/auth/v1`;

const authService = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await axios.post(`${API_BASE_URL}/login`, credentials);
      const data = response.data;
      
      if (data.token) {
        localStorage.setItem('authToken', data.token);
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

      if (data.token) {
        localStorage.setItem('authToken', data.token);
      }

      return data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(axiosError.response?.data?.message || 'Registration failed');
    }
  },

  logout: async (): Promise<void> => {
    try {
      localStorage.removeItem('authToken');
    } catch (error) {
      console.error('Logout error:', error);
      throw new Error('Logout failed');
    }
  },

  getAuthHeaders: () => {
    const token = localStorage.getItem('authToken');
    return token ? { Authorization: `Bearer ${token}` } : {};
  }
};

export default authService;

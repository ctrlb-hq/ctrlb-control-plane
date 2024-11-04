import axios, { AxiosError } from 'axios';
import { LoginCredentials, RegisterCredentials, AuthResponse, ApiError } from '../types/auth.types';

const apiUrl = import.meta.env.VITE_BACKEND_URI;
const API_BASE_URL = `${apiUrl}/api/auth/v1`;

const authService = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await axios.post(`${API_BASE_URL}/login`, credentials, {
        withCredentials: true
      });
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(axiosError.response?.data?.message || 'Login failed');
    }
  },

  register: async (userDetails: RegisterCredentials): Promise<AuthResponse> => {
    try {
      const response = await axios.post(`${API_BASE_URL}/register`, userDetails);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(axiosError.response?.data?.message || 'Registration failed');
    }
  },

  logout: async (): Promise<void> => {
    try {
      await axios.post(`${API_BASE_URL}/logout`, {}, {
        withCredentials: true,
      });
      localStorage.removeItem('sessionID');
    }  catch (error) {
      console.error('Logout error:', error);
      throw new Error('Logout failed');
    }
  }
};

export default authService;

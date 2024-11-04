import axios, { AxiosError } from 'axios';
import { Agent, ApiError } from '../types/agents.types';
import { Pipeline } from '../types/pipeline.types';

const apiUrl = import.meta.env.VITE_BACKEND_URI;
const AGENTS_BASE_URL = `${apiUrl}/api/frontend/v1/agents`;
const PIPELINES_BASE_URL = `${apiUrl}/api/frontend/v1/pipelines`;

const queryService = {

  // Agents endpoints

  fetchAgents: async (): Promise<Agent[]> => {
    try {
      const response = await axios.get<Agent[]>(AGENTS_BASE_URL);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch agents'
      );
    }
  },
  fetchAgentById: async (id: string): Promise<Agent> => {
    try {
      const response = await axios.get<Agent>(`${AGENTS_BASE_URL}/${id}`);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch agent'
      );
    }
  },
  deleteAgent: async (id: string): Promise<void> => {
    try {
      await axios.delete(`${AGENTS_BASE_URL}/${id}`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to delete agent'
      );
    }
  },
  startAgent: async (id: string): Promise<void> => {
    try {
      await axios.post(`${AGENTS_BASE_URL}/${id}/start`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to start agent'
      );
    }
  },

  stopAgent: async (id: string): Promise<void> => {
    try {
      await axios.post(`${AGENTS_BASE_URL}/${id}/stop`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to stop agent'
      );
    }
  },

  getAgentMetrics: async (id: string): Promise<unknown> => {
    try {
      const response = await axios.get(`${AGENTS_BASE_URL}/${id}/metrics`);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch agent metrics'
      );
    }
  },

  restartAgentMonitoring: async (id: string): Promise<void> => {
    try {
      await axios.post(`${AGENTS_BASE_URL}/${id}/restart-monitoring`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to restart agent monitoring'
      );
    }
  },


  // Pipelines endpoints

  fetchPipelines: async (): Promise<Pipeline[]> => {
    try {
      const response = await axios.get<Pipeline[]>(PIPELINES_BASE_URL);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch pipelines'
      );
    }
  },

  fetchPipelineById: async (id: string): Promise<Pipeline> => {
    try {
      const response = await axios.get<Pipeline>(`${PIPELINES_BASE_URL}/${id}`);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch pipeline'
      );
    }
  },

  deletePipeline: async (id: string): Promise<void> => {
    try {
      await axios.delete(`${PIPELINES_BASE_URL}/${id}`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to delete pipeline'
      );
    }
  },

  startPipeline: async (id: string): Promise<void> => {
    try {
      await axios.post(`${PIPELINES_BASE_URL}/${id}/start`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to start pipeline'
      );
    }
  },

  stopPipeline: async (id: string): Promise<void> => {
    try {
      await axios.post(`${PIPELINES_BASE_URL}/${id}/stop`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to stop pipeline'
      );
    }
  },

  getPipelineMetrics: async (id: string): Promise<unknown> => {
    try {
      const response = await axios.get(`${PIPELINES_BASE_URL}/${id}/metrics`);
      return response.data;
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to fetch pipeline metrics'
      );
    }
  },

  restartPipelineMonitoring: async (id: string): Promise<void> => {
    try {
      await axios.post(`${PIPELINES_BASE_URL}/${id}/restart-monitoring`);
    } catch (error) {
      const axiosError = error as AxiosError<ApiError>;
      throw new Error(
        axiosError.response?.data?.message || 'Failed to restart pipeline monitoring'
      );
    }
  },
};

export default queryService;

import axios, { AxiosError } from "axios";
import { ApiError } from "../types/agent.types";
import { Pipeline } from "../types/pipeline.types";
import { AgentValuesTable } from "@/types/agentValues.type";

const apiUrl = (import.meta as ImportMetaWithEnv).env.VITE_API_URL;
const AGENTS_BASE_URL = `${apiUrl}/api/frontend/v1/agents`;
const PIPELINES_BASE_URL = `${apiUrl}/api/frontend/v1/pipelines`;

const queryService = {
	// Agents endpoints

	fetchAgents: async (): Promise<AgentValuesTable[]> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get<AgentValuesTable[]>(AGENTS_BASE_URL, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch agents");
		}
	},

	fetchAgentById: async (id: string): Promise<AgentValuesTable> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get<AgentValuesTable>(`${AGENTS_BASE_URL}/${id}`, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch agent");
		}
	},

	deleteAgent: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.delete(`${AGENTS_BASE_URL}/${id}`, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to delete agent");
		}
	},

	startAgent: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${AGENTS_BASE_URL}/${id}/start`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to start agent");
		}
	},

	stopAgent: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${AGENTS_BASE_URL}/${id}/stop`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to stop agent");
		}
	},

	getAgentMetrics: async (id: string): Promise<unknown> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get(`${AGENTS_BASE_URL}/${id}/metrics`, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch agent metrics");
		}
	},

	restartAgentMonitoring: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${AGENTS_BASE_URL}/${id}/restart-monitoring`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to restart agent monitoring");
		}
	},

	// Pipelines endpoints
	fetchPipelines: async (): Promise<Pipeline[]> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get<Pipeline[]>(PIPELINES_BASE_URL, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch pipelines");
		}
	},

	fetchPipelineById: async (id: string): Promise<Pipeline> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get<Pipeline>(`${PIPELINES_BASE_URL}/${id}`, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch pipeline");
		}
	},

	deletePipeline: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.delete(`${PIPELINES_BASE_URL}/${id}`, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to delete pipeline");
		}
	},

	startPipeline: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${PIPELINES_BASE_URL}/${id}/start`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to start pipeline");
		}
	},

	stopPipeline: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${PIPELINES_BASE_URL}/${id}/stop`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to stop pipeline");
		}
	},

	getPipelineMetrics: async (id: string): Promise<unknown> => {
		try {
			const token = localStorage.getItem("authToken");
			const response = await axios.get(`${PIPELINES_BASE_URL}/${id}/metrics`, {
				headers: {
					Authorization: token,
				},
			});
			return response.data;
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to fetch pipeline metrics");
		}
	},

	restartPipelineMonitoring: async (id: string): Promise<void> => {
		try {
			const token = localStorage.getItem("authToken");
			await axios.post(`${PIPELINES_BASE_URL}/${id}/restart-monitoring`, null, {
				headers: {
					Authorization: token,
				},
			});
		} catch (error) {
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data?.message || "Failed to restart pipeline monitoring");
		}
	},
};

interface ImportMetaWithEnv extends ImportMeta {
	env: {
		VITE_API_URL: string;
	};
}

export default queryService;

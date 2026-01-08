import axiosInstance from "@/utils/axiosInstance";
import { ApiError } from "@/types/agent.types";
import { AxiosError } from "axios";

const agentServices = {
	getAgentById: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/agents/${id}`);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.getAgentById(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to fetch agent details",
			);
		}
	},
	deleteAgent: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.delete(`/agents/${id}`);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.deleteAgent(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to delete agent",
			);
		}
	},
	startAgent: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.post(`/agents/${id}/start`);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.startAgent(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to start agent",
			);
		}
	},
	stopAgent: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.post(`/agents/${id}/stop`);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.stopAgent(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to stop agent",
			);
		}
	},
	getLatestAgents: async ({ since }: { since: number }): Promise<any> => {
		try {
			const response = await axiosInstance.get("/latest-agent", {
				params: { since },
			});
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await agentServices.getLatestAgents({ since });
			}
			console.log(error);
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to fetch agent list");
		}
	},
	getAllAgents: async (): Promise<any> => {
		try {
			const response = await axiosInstance.get("/agents");
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.getAllAgents();
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to fetch all agents",
			);
		}
	},
	getAgentHealthMetrics: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/agents/${id}/healthmetrics`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await agentServices.getAgentHealthMetrics(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to get health metrics of the agent",
			);
		}
	},
	getAgentRateMetrics: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/agents/${id}/ratemetrics`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await agentServices.getAgentRateMetrics(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to get Rate metrics of the agent");
		}
	},
	addAgentLabels: async (
		id: number,
		labels: Record<string, string>
	): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.post(`/agents/${id}/labels`, {
				labels,
			});
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.addAgentLabels(id, labels);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to add labels to agent",
			);
		}
	},
	attachAgentToPipeline: async (
		pipelineId: number,
		agentId: number
	): Promise<any> => {
		try {
			if (!pipelineId || !agentId) return;
			const response = await axiosInstance.post(
				`/pipelines/${pipelineId}/agents/${agentId}`
			);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.attachAgentToPipeline(
					pipelineId,
					agentId
				);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message ||
					"Failed to attach agent to pipeline",
			);
		}
	},
	detachAgentFromPipeline: async (
		pipelineId: number,
		agentId: number
	): Promise<any> => {
		try {
			if (!pipelineId || !agentId) return;
			const response = await axiosInstance.delete(
				`/pipelines/${pipelineId}/agents/${agentId}`
			);
			return response.data;
		} catch (error: any) {
			if (error.response?.status === 401) {
				return await agentServices.detachAgentFromPipeline(
					pipelineId,
					agentId
				);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message ||
					"Failed to detach agent from pipeline",
			);
		}
	},
};

export default agentServices;

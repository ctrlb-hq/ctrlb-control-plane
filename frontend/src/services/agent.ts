import axiosInstance from "@/utils/axiosInstance";
import { ApiError } from "@/types/agent.types";
import { AxiosError } from "axios";

const agentServices = {
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
	restartAgentMonitoring: async (id: number): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.post(`/agents/${id}/restart-monitoring`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await agentServices.restartAgentMonitoring(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message || "Failed to restart monitoring of the agent",
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
};

export default agentServices;

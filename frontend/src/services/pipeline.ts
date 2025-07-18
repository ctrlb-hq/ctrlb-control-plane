import axiosInstance from "@/utils/axiosInstance";
import { ApiError } from "@/types/agent.types";
import { AxiosError } from "axios";

const pipelineServices = {
	getAllPipelines: async (): Promise<any> => {
		try {
			const response = await axiosInstance.get(`/pipelines`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.getAllPipelines();
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to fetch pipelines list");
		}
	},

	getPipelineById: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/pipelines/${id}`);
			const data = response.data;
			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.getPipelineById(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to fetch Pipeline by it's Id.");
		}
	},
	getPipelineOverviewById: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/pipelines-overview/${id}`);
			const data = response.data;
			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.getPipelineOverviewById(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to fetch Pipeline by it's Id.");
		}
	},
	deletePipelineById: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			await axiosInstance.delete(`/pipelines/${id}`);
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.deletePipelineById(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to delete pipeline by it's id");
		}
	},
	getPipelineGraph: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/pipelines/${id}/graph`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.getPipelineGraph(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to get pipeline graph.");
		}
	},
	getAllAgentsAttachedToPipeline: async (id: string): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.get(`/pipelines/${id}/agents`);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.getAllAgentsAttachedToPipeline(id);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(
				axiosError.response?.data.message ||
					"Failed to fetch agents connected to the given pipeline by it's id.",
			);
		}
	},
	syncPipelineGraph: async (id: string, payload: any): Promise<any> => {
		try {
			if (!id) return;
			const response = await axiosInstance.post(`/pipelines/${id}/graph`, payload);
			const data = response.data;

			return data;
		} catch (error: any) {
			if (error.response.status === 401) {
				return await pipelineServices.syncPipelineGraph(id, payload);
			}
			const axiosError = error as AxiosError<ApiError>;
			throw new Error(axiosError.response?.data.message || "Failed to sync pipeline graph.");
		}
	},
};

export default pipelineServices;

import axiosInstance from '@/lib/axiosInstance';
import { ApiError } from '@/types/agent.types';
import { AxiosError } from 'axios';

const pipelineServices = {
    getAllPipelines: async (): Promise<any> => {
        try {
            const response = await axiosInstance.get(`/pipelines`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.getAllPipelines()
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch pipelines list")
        }
    },

    addPipeline: async (payload: any): Promise<any> => {
        try {
            const response = await axiosInstance.post(`/pipelines`, payload)
            const data = response.data
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.addPipeline(payload)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to add pipeline.")
        }
    },

    getPipelineById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/pipelines/${id}`)
            const data = response.data
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.getPipelineById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch Pipeline by it's Id.")
        }
    },
    deletePipelineById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            await axiosInstance.delete(`/pipelines/${id}`)
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.deletePipelineById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to delete pipeline by it's id")
        }
    },
    getPipelineGraph: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/pipelines/${id}/graph`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.getPipelineGraph(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to get pipeline graph.")
        }
    },
    syncPipelineGraph: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.post(`/pipelines/${id}/graph`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.syncPipelineGraph(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to sync pipeline graph.")
        }
    },
    getAllAgentsAttachedToPipeline: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/pipelines/${id}/agents`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.getAllAgentsAttachedToPipeline(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch agents connected to the given pipeline by it's id.")
        }
    },
    detachAgentFromPipeline: async (id: string, agent_id: string): Promise<any> => {
        try {
            if (!id || !agent_id) return
            await axiosInstance.delete(`/pipelines/${id}/agents/${agent_id}`)

        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.detachAgentFromPipeline(id, agent_id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to detach agent from the pipeline.")
        }
    },
    attachAgentToPipeline: async (id: string, agent_id: string): Promise<any> => {
        try {
            if (!id || !agent_id) return
            const response = await axiosInstance.post(`/pipelines/${id}/agents/${agent_id}`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.attachAgentToPipeline(id, agent_id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to attach an agent to the pipeline.")
        }
    },
    getAllUnattachedAgents: async (): Promise<any> => {
        try {
            const response = await axiosInstance.get(`/unassigned-agents`)
            const data = response.data

            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await pipelineServices.getAllUnattachedAgents(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch unattached agents.")
        }
    }
}



export default pipelineServices
import axiosInstance from '@/lib/axiosInstance';
import { ApiError } from '@/types/agent.types';
import { AxiosError } from 'axios';

const agentServices = {
    getAllAgents: async (): Promise<any> => {
        try {
            const response = await axiosInstance.get("/agents")
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.getAllAgents()
            }
            console.log(error)
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch agent list")
        }
    },
    getLatestAgents: async ({ since }: { since: number }): Promise<any> => {
        try {
            const response = await axiosInstance.get("/latest‑agent", {
                params: { since }
            })
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.getLatestAgents({ since })
            }
            console.log(error)
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch agent list")
        }
    },

    getAgentById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/agents/${id}`)
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.getAgentById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch agent by its Id.")
        }
    },

    deleteAgentById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.delete(`/agents/${id}`)
            const data = response.data
            
            console.log("Agent Deleted successfully")
            return data
        } catch (error: any) {
            if (error.response.status === 401) {              
                return await agentServices.deleteAgentById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to delete agent by its id")
        }
    },

    startAgentById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.post(`/agents/${id}/start`)
            const data = response.data
            
            console.log("Agent Started successfully")
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.startAgentById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to start agent by its id")
        }
    },

    stopAgentById: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.post(`/agents/${id}/stop`)
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.stopAgentById(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to stop agent by its id")
        }
    },

    restartAgentMonitoring: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.post(`/agents/${id}/restart-monitoring`)
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.restartAgentMonitoring(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to restart monitoring of the agent")
        }
    },

    getAgentHealthMetrics: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/agents/${id}/healthmetrics`)
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.getAgentHealthMetrics(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to get health metrics of the agent")
        }
    },
    getAgentRateMetrics: async (id: string): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.get(`/agents/${id}/ratemetrics`)
            const data = response.data
            
            return data
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.getAgentRateMetrics(id)
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to get Rate metrics of the agent")
        }
    },
    addAgentLabel: async (id: string, label: { [key: string]: string }): Promise<any> => {
        try {
            if (!id) return
            const response = await axiosInstance.post(`/agents/${id}/labels`, label);
            const data = response.data;

            return data;
        } catch (error: any) {
            if (error.response.status === 401) {
                return await agentServices.addAgentLabel(id, label);
            }
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to add label");
        }
    },
}



export default agentServices
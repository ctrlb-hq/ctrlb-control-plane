import axiosInstance from '@/utils/axiosInstance';
import { ApiError } from '@/types/agent.types';
import { AxiosError } from 'axios';

export const TransporterService = {
    getTransporterService: async (type: string): Promise<any> => {
        try {
            const res = await axiosInstance.get('/component', {
                params: {
                    type: type
                }
            })
            const data = res.data
            return data

        } catch (error: any) {
            if (error.response.status === 401) {
                return await TransporterService.getTransporterService(type)
            }
            console.log(error)
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch get transport options.")
        }
    },

    getTransporterForm: async (name: string): Promise<any> => {
        try {
            const res=await axiosInstance.get(`/component/schema/${name}`)
            const data=res.data
            return data
        } catch (error:any) {
            if (error.response.status === 401) {
                return await TransporterService.getTransporterService(name)
            }
            console.log(error)
            const axiosError = error as AxiosError<ApiError>;
            throw new Error(axiosError.response?.data.message || "Failed to fetch get transport form.")
        }
    }
}
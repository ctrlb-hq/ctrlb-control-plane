import { formatTimestampWithDate } from '@/constants';
import { useToast } from '@/hooks/use-toast';
import agentServices from '@/services/agentServices';
import pipelineServices from '@/services/pipelineServices';
import { RefreshCw } from 'lucide-react';
import { useEffect, useState } from 'react'

const PipelineOverview = ({pipelineId}:{pipelineId:string}) => {
    const [pipelineOverviewData, setPipelineOverviewData] = useState<any>(null);
    const { toast } = useToast()

    const handleGetPipelineOverview = async () => {
        try {
            const response = await pipelineServices.getPipelineOverviewById(pipelineId);
            setPipelineOverviewData(response);
        } catch (error) {
            console.error("Error fetching pipeline overview:", error);
            toast({
                title: "Error",
                description: "Failed to fetch pipeline overview",
                variant: "destructive",
            });
        }
    };


    const handleRefreshStatus = async () => {
        try {
            if (!pipelineOverviewData?.agent_id) return;
            await agentServices.restartAgentMonitoring(pipelineOverviewData.agent_id);
            // Refresh the pipeline data using the existing function
            await handleGetPipelineOverview();
            toast({
                title: "Success",
                description: "Pipeline status refreshed successfully",
            });
        } catch (error) {
            console.error("Failed to refresh pipeline status:", error);
            toast({
                title: "Error",
                description: "Failed to refresh pipeline status",
                variant: "destructive",
            });
        }
    };

    useEffect(() => {
        handleGetPipelineOverview();
    }, [pipelineId])


    return (
        <>
            <div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm px-4 py-2 mb-2">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-3 text-gray-700 text-sm">
                    <div>
                        <p className="text-gray-500 leading-tight">Name</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.name || "-"}</p>
                    </div>
                    <div className="flex flex-col">
                        <p className="text-gray-500 leading-tight">Status</p>
                        <div className="flex items-center gap-2">
                            <span
                                className={`capitalize px-2 py-0.5 rounded-full text-xs font-semibold ${pipelineOverviewData?.status?.toLowerCase() === "connected"
                                    ? "bg-green-200 text-green-700"
                                    : pipelineOverviewData?.status?.toLowerCase() === "disconnected"
                                        ? "bg-red-100 text-red-700"
                                        : "bg-yellow-100 text-yellow-700"
                                    }`}>
                                {pipelineOverviewData?.status}
                            </span>
                            {["disconnected", "pending", "inactive"].includes(
                                pipelineOverviewData?.status?.toLowerCase(),
                            ) && (
                                    <RefreshCw
                                        className="h-3.5 w-3.5 text-gray-500 cursor-pointer hover:text-gray-700 transition-transform hover:rotate-180"
                                        onClick={handleRefreshStatus}
                                    />
                                )}
                        </div>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Created At</p>
                        <p className="font-medium leading-tight">{formatTimestampWithDate(pipelineOverviewData?.created_at)}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Created By</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.created_by || "-"}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Updated At</p>
                        <p className="font-medium leading-tight">{formatTimestampWithDate(pipelineOverviewData?.updated_at)}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Hostname</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.hostname}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Agent Version</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.agent_version}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">IP Address</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.ip_address}</p>
                    </div>
                    <div>
                        <p className="text-gray-500 leading-tight">Platform</p>
                        <p className="font-medium leading-tight">{pipelineOverviewData?.platform}</p>
                    </div>
                </div>
            </div>
        </>
    )
}

export default PipelineOverview

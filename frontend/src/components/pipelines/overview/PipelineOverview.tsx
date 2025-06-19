import { formatTimestampWithDate } from "@/constants";
import { useToast } from "@/hooks/useToast";
import agentServices from "@/services/agent";
import pipelineServices from "@/services/pipeline";
import { MetricData } from "@/types/pipeline.types";
import { RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";

import { HealthChart } from "./HealthChart";
import { getRandomChartColor } from "@/constants";

type Props = {
	pipelineId: string;
};

const PipelineOverview = ({ pipelineId }: Props) => {
	const [pipelineOverviewData, setPipelineOverviewData] = useState<any>(null);
	const [healthMetrics, setHealthMetrics] = useState<MetricData[]>([]);

	const { toast } = useToast();

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
	const fetchHealthMetrics = async () => {
		try {
			const metrics = await agentServices.getAgentHealthMetrics(pipelineOverviewData.agent_id);
			if (
				Array.isArray(metrics) &&
				metrics.length > 0 &&
				metrics.every(
					metric => metric?.data_points && Array.isArray(metric.data_points) && metric.metric_name,
				)
			) {
				setHealthMetrics(metrics);
			} else {
				setHealthMetrics([]); // Set empty array for invalid/null data
			}
		} catch (error) {
			console.error("Error fetching health metrics:", error);
			toast({
				title: "Error",
				description: error instanceof Error ? error.message : "Failed to fetch health metrics",
				variant: "destructive",
			});
			// Set empty array instead of leaving previous state
			setHealthMetrics([]);
		}
	};

	useEffect(() => {
		handleGetPipelineOverview();
	}, [pipelineId]);

	useEffect(() => {
		if (pipelineOverviewData) {
			fetchHealthMetrics();
		}
	}, [pipelineOverviewData]);

	return (
		<>
			<div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm px-4 py-2 mb-2">
				<div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-3 text-gray-700 text-sm">
					<div>
						<p className="text-gray-500 leading-tight">Pipeline Name</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.name || "-"}</p>
					</div>
					<div className="flex flex-col">
						<p className="text-gray-500 leading-tight">Pipeline ID</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.id || "-"}</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Created At</p>
						<p className="font-medium leading-tight">
							{formatTimestampWithDate(pipelineOverviewData?.created_at)}
						</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Created By</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.created_by || "-"}</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Updated At</p>
						<p className="font-medium leading-tight">
							{formatTimestampWithDate(pipelineOverviewData?.updated_at)}
						</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Status</p>
						<div className="flex items-center gap-2">
							<span
								className={`capitalize px-2 py-0.5 rounded-full text-xs font-semibold ${
									pipelineOverviewData?.status?.toLowerCase() === "connected"
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
						<p className="text-gray-500 leading-tight">Collector Version</p>
						<p className="font-medium leading-tight">v{pipelineOverviewData?.agent_version}</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Hostname</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.hostname}</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">Platform</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.platform}</p>
					</div>
					<div>
						<p className="text-gray-500 leading-tight">IP Address</p>
						<p className="font-medium leading-tight">{pipelineOverviewData?.ip_address}</p>
					</div>
				</div>
			</div>
			<div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
				{healthMetrics.length > 0 ? (
					healthMetrics.map(metric => (
						<div key={metric.metric_name} className="w-full h-[150px] bg-white rounded-lg shadow-sm">
							<HealthChart
								name={metric.metric_name === "cpu_utilization" ? "CPU Usage" : "Memory Usage"}
								data={metric.data_points.map(point => ({
									timestamp: point.timestamp,
									[metric.metric_name]:
										metric.metric_name === "memory_utilization" ? point.value / (1024 * 1024) : point.value,
								}))}
								y_axis_data_key={metric.metric_name}
								chart_color={getRandomChartColor(metric.metric_name)}
								yAxisLabel={
									metric.metric_name === "cpu_utilization" ? "CPU Utilization (%)" : "Memory Utilization (%)"
								}
							/>
						</div>
					))
				) : (
					<div className="col-span-2 bg-white rounded-lg shadow-sm flex flex-col items-center justify-center min-h-[120px]">
						<div className="text-gray-400 mb-2">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								className="h-8 w-8"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor">
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									strokeWidth={2}
									d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
								/>
							</svg>
						</div>
						<p className="text-gray-500 text-base font-medium">No Health Metrics Available</p>
						<p className="text-gray-400 text-xs mt-1">
							Health metrics will appear here once data is available
						</p>
					</div>
				)}
			</div>
		</>
	);
};

export default PipelineOverview;

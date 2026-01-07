import { formatTimestampWithDate } from "@/constants";
import agentServices from "@/services/agent";
import { MetricData } from "@/types/pipeline.types";
import { RefreshCw, Loader2 } from "lucide-react";
import { useEffect, useState, useCallback } from "react";
import { HealthChart } from "./HealthChart";
import { getRandomChartColor } from "@/constants";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import { PipelineOverviewInterface } from "@/types/pipeline.types";

type Props = {
	pipelineOverviewData: PipelineOverviewInterface;
	onRefresh: () => Promise<void>;
};

const PipelineOverview = ({ pipelineOverviewData, onRefresh }: Props) => {
	const [healthMetrics, setHealthMetrics] = useState<MetricData[]>([]);
	const [metricsLoading, setMetricsLoading] = useState(false);
	const { showSnackbar } = useGlobalSnackbar();

	const handleRefreshStatus = async () => {
		try {
			if (!pipelineOverviewData?.agent_id) return;

			await agentServices.restartAgentMonitoring(pipelineOverviewData.agent_id);
			await onRefresh();

			showSnackbar("Pipeline status refresh initiated", "success");
		} catch (error) {
			console.error("Failed to refresh pipeline status:", error);
			showSnackbar("Pipeline status refresh failed", "error");
		}
	};

	const fetchHealthMetrics = useCallback(async () => {
		if (!pipelineOverviewData?.agent_id) return;

		setMetricsLoading(true);
		setHealthMetrics([]);

		try {
			const metrics = await agentServices.getAgentHealthMetrics(
				pipelineOverviewData.agent_id
			);

			if (
				Array.isArray(metrics) &&
				metrics.length > 0 &&
				metrics.every(
					m => m?.data_points && Array.isArray(m.data_points) && m.metric_name
				)
			) {
				setHealthMetrics(metrics);
			}
		} catch (error) {
			console.error("Error fetching health metrics:", error);
			showSnackbar(
				error instanceof Error ? error.message : "Failed to fetch health metrics",
				"error"
			);
		} finally {
			setMetricsLoading(false);
		}
	}, [pipelineOverviewData?.agent_id, showSnackbar]);

	useEffect(() => {
		if (!pipelineOverviewData?.agent_id) return;
		fetchHealthMetrics();
	}, [pipelineOverviewData?.agent_id, fetchHealthMetrics]);

	return (
		<>
			<div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm px-4 py-2 mb-2">
				<div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-3 text-gray-700 text-sm">
					<div>
						<p className="text-gray-500">Pipeline Name</p>
						<p className="font-medium">{pipelineOverviewData.name}</p>
					</div>
					<div>
						<p className="text-gray-500">Pipeline ID</p>
						<p className="font-medium">{pipelineOverviewData.id}</p>
					</div>
					<div>
						<p className="text-gray-500">Created At</p>
						<p className="font-medium">
							{formatTimestampWithDate(pipelineOverviewData.created_at)}
						</p>
					</div>
					<div>
						<p className="text-gray-500">Created By</p>
						<p className="font-medium">
							{pipelineOverviewData.created_by || "-"}
						</p>
					</div>
					<div>
						<p className="text-gray-500">Updated At</p>
						<p className="font-medium">
							{formatTimestampWithDate(pipelineOverviewData.updated_at)}
						</p>
					</div>
					<div>
						<p className="text-gray-500">Status</p>
						<div className="flex items-center gap-2">
							<span
								className={`capitalize px-2 py-0.5 rounded-full text-xs font-semibold ${
									pipelineOverviewData.status?.toLowerCase() === "connected"
										? "bg-green-200 text-green-700"
										: pipelineOverviewData.status?.toLowerCase() === "disconnected"
										? "bg-red-100 text-red-700"
										: "bg-yellow-100 text-yellow-700"
								}`}
							>
								{pipelineOverviewData.status}
							</span>
							{["disconnected", "pending", "inactive"].includes(
								pipelineOverviewData.status?.toLowerCase()
							) && (
								<RefreshCw
									className="h-3.5 w-3.5 cursor-pointer"
									onClick={handleRefreshStatus}
								/>
							)}
						</div>
					</div>
					<div>
						<p className="text-gray-500">Collector Version</p>
						<p className="font-medium">
							v{pipelineOverviewData.agent_version}
						</p>
					</div>
					<div>
						<p className="text-gray-500">Hostname</p>
						<p className="font-medium">{pipelineOverviewData.hostname}</p>
					</div>
					<div>
						<p className="text-gray-500">Platform</p>
						<p className="font-medium">{pipelineOverviewData.platform}</p>
					</div>
					<div>
						<p className="text-gray-500">IP Address</p>
						<p className="font-medium">{pipelineOverviewData.ip_address}</p>
					</div>
				</div>
			</div>
			<div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
				{metricsLoading ? (
					<div className="col-span-2 flex justify-center items-center min-h-[150px]">
						<Loader2 className="h-6 w-6 animate-spin text-gray-500" />
					</div>
				) : healthMetrics.length > 0 ? (
					healthMetrics.map(metric => (
						<div
							key={metric.metric_name}
							className="w-full h-[150px] bg-white rounded-lg shadow-sm"
						>
							<HealthChart
								name={
									metric.metric_name === "cpu_utilization"
										? "CPU Usage"
										: "Memory Usage"
								}
								data={metric.data_points.map(point => ({
									timestamp: point.timestamp,
									[metric.metric_name]:
										metric.metric_name === "memory_utilization"
											? point.value / (1024 * 1024)
											: point.value,
								}))}
								y_axis_data_key={metric.metric_name}
								chart_color={getRandomChartColor(metric.metric_name)}
								yAxisLabel={
									metric.metric_name === "cpu_utilization"
										? "CPU Utilization (%)"
										: "Memory Utilization (MB)"
								}
							/>
						</div>
					))
				) : (
					<div className="col-span-2 bg-white rounded-lg shadow-sm flex flex-col items-center justify-center min-h-[120px]">
						<p className="text-gray-500 font-medium">
							No Health Metrics Available
						</p>
						<p className="text-gray-400 text-xs mt-1">
							Health metrics will appear once data is available
						</p>
					</div>
				)}
			</div>
		</>
	);
};

export default PipelineOverview;

import {
	Drawer,
	Box,
	Typography,
	Divider,
	IconButton,
	Tooltip,
} from "@mui/material";
import { X, Play, Square, Loader2 } from "lucide-react";
import { useEffect, useState, useCallback, useRef } from "react";
import LinkIcon from "@mui/icons-material/Link";
import LinkOffIcon from "@mui/icons-material/LinkOff";
import agentServices from "@/services/agent";
import { MetricData } from "@/types/pipeline.types";
import { getRandomChartColor } from "@/constants";
import { HealthChart } from "@/components/pipelines/overview/HealthChart";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import AttachAgentDialog from "./AttachAgentDialog";
import DetachAgentDialog from "./DetachAgentDialog";
import { agentVal } from "@/types/agent.types";

interface ViewAgentDetailsProps {
	agentId: number;
	open: boolean;
	onClose: () => void;
}

const MIN_WIDTH = 800;
const MAX_WIDTH = 1200;

const ViewAgentDetails = ({
	agentId,
	open,
	onClose,
}: ViewAgentDetailsProps) => {
	const [drawerWidth, setDrawerWidth] = useState(820);
	const [healthMetrics, setHealthMetrics] = useState<MetricData[]>([]);
	const [metricsLoading, setMetricsLoading] = useState(false);
	const [agentDetails, setAgentDetails] = useState<agentVal | null>(null);
	const [attachOpen, setAttachOpen] = useState(false);
	const [detachOpen, setDetachOpen] = useState(false);

	const isResizing = useRef(false);
	const { showSnackbar } = useGlobalSnackbar();

	const handleMouseDown = () => {
		isResizing.current = true;
		document.body.style.cursor = "col-resize";
	};

	const handleMouseMove = (e: MouseEvent) => {
		if (!isResizing.current) return;

		const newWidth = window.innerWidth - e.clientX;
		if (newWidth >= MIN_WIDTH && newWidth <= MAX_WIDTH) {
			setDrawerWidth(newWidth);
		}
	};

	const handleMouseUp = () => {
		isResizing.current = false;
		document.body.style.cursor = "default";
	};

	useEffect(() => {
		window.addEventListener("mousemove", handleMouseMove);
		window.addEventListener("mouseup", handleMouseUp);

		return () => {
			window.removeEventListener("mousemove", handleMouseMove);
			window.removeEventListener("mouseup", handleMouseUp);
		};
	}, []);

	const fetchAgentDetails = useCallback(async () => {
		if (!agentId) return;
		try {
			const data = await agentServices.getAgentById(agentId);
			setAgentDetails(data);
		} catch {
			showSnackbar("Failed to fetch agent details", "error");
		}
	}, [agentId, showSnackbar]);

	const fetchHealthMetrics = useCallback(async () => {
		if (!agentId) return;

		setMetricsLoading(true);
		setHealthMetrics([]);

		try {
			const metrics = await agentServices.getAgentHealthMetrics(agentId);
			if (Array.isArray(metrics)) setHealthMetrics(metrics);
		} catch (error) {
			showSnackbar(
				error instanceof Error ? error.message : "Failed to fetch health metrics",
				"error"
			);
		} finally {
			setMetricsLoading(false);
		}
	}, [agentId, showSnackbar]);

	const handleStart = async () => {
		try {
			await agentServices.startAgent(agentId);
			showSnackbar("Agent start initiated", "success");
			await fetchAgentDetails();
		} catch {
			showSnackbar("Failed to start agent", "error");
		}
	};

	const handleStop = async () => {
		try {
			await agentServices.stopAgent(agentId);
			showSnackbar("Agent stop initiated", "success");
			await fetchAgentDetails();
		} catch {
			showSnackbar("Failed to stop agent", "error");
		}
	};

	useEffect(() => {
		if (!open) return;
		fetchAgentDetails();
		fetchHealthMetrics();
	}, [open, fetchAgentDetails, fetchHealthMetrics]);

	const pipelineId =
		agentDetails?.pipeline_id
			? Number(agentDetails.pipeline_id)
			: null;

	const isAttached = Boolean(pipelineId);


	return (
		<>
			<Drawer
				anchor="right"
				open={open}
				onClose={onClose}
				PaperProps={{
				sx: {
					display: "flex",
					flexDirection: "row",
				},
				}}
			>
				<div
					onMouseDown={handleMouseDown}
					style={{
						width: 10,
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						cursor: "col-resize",
					}}
				>
					<div
						style={{
							width: 4,
							height: 48,
							borderRadius: 2,
							backgroundColor: "#93c5fd",
						}}
					/>
				</div>
				<Box
					sx={{
						width: drawerWidth,
						p: 3,
						height: "100%",
						overflowY: "auto",
					}}
				>
					{/* Header */}
					<Box className="flex justify-between items-center mb-2">
						<Typography variant="h6" fontWeight={600}>
							{agentDetails?.name ?? "Agent" }
						</Typography>

						<Box className="flex items-center gap-1">
							<Box
								sx={{
									display: "flex",
									alignItems: "center",
									border: "1px solid",
									borderColor: "grey.300",
									borderRadius: 1,
									overflow: "hidden",
									marginRight: 2,
								}}
							>
								{/* Attach */}
								<Tooltip title="Attach agent to pipeline" arrow>
									<IconButton
										size="small"
										onClick={() => setAttachOpen(true)}
										sx={{
											color: "primary.main",
											borderRadius: 0,
											"&:hover": {
												backgroundColor: "rgba(59,130,246,0.1)",
											},
										}}
									>
										<LinkIcon fontSize="small" />
									</IconButton>
								</Tooltip>

								{/* Divider */}
								<Box
									sx={{
										width: "1px",
										height: 24,
										backgroundColor: "grey.300",
									}}
								/>

								{/* Detach */}
								<Tooltip
									title={
										isAttached
											? "Detach agent from pipeline"
											: "No pipeline attached"
									}
									arrow
								>
									<span>
										<IconButton
											size="small"
											color="error"
											disabled={!isAttached}
											onClick={() => {
												if (isAttached) setDetachOpen(true);
											}}
											sx={{
												borderRadius: 0,
												"&:hover": {
													backgroundColor: isAttached
														? "rgba(239,68,68,0.1)"
														: "transparent",
												},
											}}
										>
											<LinkOffIcon fontSize="small" />
										</IconButton>
									</span>
								</Tooltip>
							</Box>

							{/* Close */}
							<IconButton
								onClick={onClose}
								sx={{
									p: 0.8,
									borderRadius: "10px",
									color: "grey.600",
									"&:hover": {
										backgroundColor: "error.main",
										color: "#fff",
									},
								}}
							>
								<X size={18} />
							</IconButton>

						</Box>
					</Box>

					<Divider className="mb-3" />

					{/* Agent Overview Card */}
					<div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm px-4 py-3 mb-3">
						<div className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
							<div>
								<p className="text-gray-500">Agent Name</p>
								<p className="font-medium">{agentDetails?.name ?? "-"}</p>
							</div>

							<div>
								<p className="text-gray-500">Agent ID</p>
								<p className="font-medium">{agentDetails?.id ?? "-"}</p>
							</div>

							<div>
								<p className="text-gray-500">Version</p>
								<p className="font-medium">
									{agentDetails?.version ?? "-"}
								</p>
							</div>

							<div>
								<p className="text-gray-500">Pipeline</p>
								<p className="font-medium">
									{agentDetails?.pipeline_name || "Not attached"}
								</p>
							</div>

							<div>
								<p className="text-gray-500">Hostname</p>
								<p className="font-medium">{agentDetails?.hostname ?? "-"}</p>
							</div>

							<div>
								<p className="text-gray-500">IP Address</p>
								<p className="font-medium">{agentDetails?.ip ?? "-"}</p>
							</div>

							<div>
								<p className="text-gray-500">Platform</p>
								<p className="font-medium">{agentDetails?.platform ?? "-"}</p>
							</div>

							{/* Status + Actions */}
							<div>
								<p className="text-gray-500 mb-1">Status</p>
								<div className="flex items-center gap-3">
									<span
										className={`capitalize px-2 py-0.5 rounded-full text-xs font-semibold ${
											agentDetails?.status === "connected"
												? "bg-green-200 text-green-700"
												: "bg-red-100 text-red-700"
										}`}
									>
										{agentDetails?.status ?? "-"}
									</span>

									{/* Start */}
									<IconButton size="small" onClick={handleStart}>
										<Play className="h-4 w-4 text-green-600" />
									</IconButton>

									{/* Stop */}
									<IconButton size="small" onClick={handleStop}>
										<Square className="h-4 w-4 text-red-600" />
									</IconButton>
								</div>
							</div>
						</div>
					</div>

					{/* Health Metrics */}
					<div className="grid grid-cols-2 gap-4">
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
				</Box>
			</Drawer>
			<AttachAgentDialog
				open={attachOpen}
				agentId={agentId}
				onClose={() => setAttachOpen(false)}
				onSuccess={() => {
					fetchAgentDetails();
				}}
			/>
			<DetachAgentDialog
				open={detachOpen}
				agentId={agentId}
				pipelineId={pipelineId ?? 0}
				onClose={() => setDetachOpen(false)}
				onSuccess={() => {
					fetchAgentDetails();
				}}
			/>
		</>
	);
};

export default ViewAgentDetails;

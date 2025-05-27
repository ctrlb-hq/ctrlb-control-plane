import { Boxes, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { useToast } from "@/hooks/use-toast";
import agentServices from "@/services/agentServices";
import pipelineServices from "@/services/pipelineServices";
import { Pipeline } from "@/types/pipeline.types";

import { Connection, Edge, EdgeMouseHandler, ReactFlowInstance } from "reactflow";
import "reactflow/dist/style.css";
import { HealthChart } from "../charts/HealthChart";
import { DestinationNode } from "./Nodes/DestinationNode";
import { ProcessorNode } from "./Nodes/ProcessorNode";
import { SourceNode } from "./Nodes/SourceNode";
import PipelineGraphEditor from "./PipelineGraphEditor";
import DeletePipelineDialog from "./DeletePipelineDialog";
interface DataPoint {
	timestamp: number;
	value: number;
}

interface MetricData {
	metric_name: string;
	data_points: DataPoint[];
}

interface FormSchema {
	title?: string;
	type?: string;
	properties?: Record<string, any>;
	required?: string[];
	[key: string]: any;
}

const statusColors: Record<string, string> = {
	connected: "text-green-600",
	disconnected: "text-red-600",
	pending: "text-yellow-600",
	inactive: "text-blue-600",
	default: "text-gray-600",
};



const ViewPipelineDetails = ({ pipelineId }: { pipelineId: string }) => {
	const {
		nodeValue,
		setNodeValueDirect,
		edgeValue,
		setEdgeValueDirect,
		updateNodes,
		updateEdges,
		connectNodes,
		resetGraph,
	} = useGraphFlow();
	const reactFlowWrapper = useRef<HTMLDivElement>(null);
	const [_reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
	const [isEditMode, setIsEditMode] = useState(false);
	const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
	const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });
	const { changesLog, clearChangesLog, addChange } = usePipelineChangesLog();
	const [pipelineOverview, setPipelineOverview] = useState<Pipeline>();
	const [isOpen, setIsOpen] = useState(false);
	const [pipelineOverviewData, setPipelineOverviewData] = useState<any>(null);
	const [healthMetrics, setHealthMetrics] = useState<MetricData[]>([]);
	const { toast } = useToast();
	const [hasDeployError, setHasDeployError] = useState(false);
	const [tabs, setTabs] = useState<string>("overview");
	const [isReviewSheetOpen, setIsReviewSheetOpen] = useState(false);
	const [isEditFormOpen, setIsEditFormOpen] = useState(false);
	const [form, setForm] = useState<FormSchema>({});
	const [config, setConfig] = useState<object>({});
	const [selectedChange, setSelectedChange] = useState<any>(null)

	console.log("xx",healthMetrics);
	const nodeTypes = useMemo(
		() => ({
			source: SourceNode,
			processor: ProcessorNode,
			destination: DestinationNode,
		}),
		[],
	);

	const handleGetPipeline = async () => {
		const res = await pipelineServices.getPipelineById(pipelineId);
		setPipelineOverview(res);
	};

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
	useEffect(() => {
		handleGetPipelineOverview();
	}, [pipelineId]);

	const handleGetPipelineGraph = async () => {
		const res = await pipelineServices.getPipelineGraph(pipelineId);
		const edges = res.edges;
		const VERTICAL_SPACING = 100;

		const updatedNodes = res.nodes.map((node: any, index: number) => {
			const nodeType =
				node.component_role === "receiver"
					? "source"
					: node.component_role === "exporter"
						? "destination"
						: "processor";

			// Calculate position based on node type
			let x, y;
			if (nodeType === "source") {
				x = 50; // Fixed left position
				y = 100 + index * VERTICAL_SPACING;
			} else if (nodeType === "destination") {
				x = 400; // Fixed right position
				y = 100 + index * VERTICAL_SPACING;
			} else {
				// processor
				x = 225; // Center position
				y = 100 + index * VERTICAL_SPACING;
			}

			return {
				id: node.component_id.toString(),
				type: nodeType,
				position: { x, y },
				data: {
					component_id: node.component_id.toString(),
					name: node.name,
					component_name: node.component_name,
					supported_signals: node.supported_signals,
					config: node.config,
				},
			};
		});
		setNodeValueDirect(updatedNodes);

		const updatedEdges = edges.map((edge: any) => ({
			id: `edge-${edge.source}-${edge.target}`,
			animated: true,
			source: edge.source,
			target: edge.target,
			data: {
				sourceComponentId: edge.source,
				targetComponentId: edge.target,
			},
		}));
		setEdgeValueDirect(updatedEdges);
	};

	useEffect(() => {
		handleGetPipelineGraph();
		handleGetPipelineOverview();
	}, [pipelineId]);

	useEffect(() => {
		handleGetPipeline();
	}, [isEditMode]);

	const onConnect = useCallback(
		(params: Edge | Connection) => {
			connectNodes(params);
		},
		[connectNodes],
	);


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
		if (pipelineOverviewData && !isEditMode) {
			fetchHealthMetrics();
			// Optional: Set up polling to refresh data periodically
			// const interval = setInterval(fetchHealthMetrics, 300000); // every 30 seconds
			// return () => clearInterval(interval);
		}
	}, [pipelineOverviewData, isEditMode]);

	const onEdgeClick: EdgeMouseHandler = useCallback(
		(event, edge) => {
			if (!isEditMode) return;
			// Calculate the position for the popover
			const rect = reactFlowWrapper.current?.getBoundingClientRect();
			if (rect) {
				setEdgePopoverPosition({
					x: event.clientX - rect.left,
					y: event.clientY - rect.top,
				});
			}

			setSelectedEdge(edge);
		},
		[isEditMode],
	);

	const handleDeleteEdge = useCallback(() => {
		if (selectedEdge) {
			const sourceNode = nodeValue.find(node => node.id === selectedEdge.source);
			const targetNode = nodeValue.find(node => node.id === selectedEdge.target);

			//Add to changes log
			const changeLogEntry = {
				type: "Connection",
				name: `${sourceNode?.data.name || "Unknown"} → ${targetNode?.data.name || "Unknown"}`,
				status: "deleted",
			};
			changesLog.push(changeLogEntry);
			//Filter out only the specific edge that matches both source and target
			const newEdges = edgeValue.filter(
				edge => !(edge.source === selectedEdge.source && edge.target === selectedEdge.target),
			);
			setEdgeValueDirect(newEdges);
			setSelectedEdge(null);
		}
	}, [selectedEdge, edgeValue, setEdgeValueDirect]);

	const onPaneClick = useCallback(() => {
		setSelectedEdge(null);
	}, []);

	const handleDeployChanges = async () => {
		try {
			const syncPayload = {
				nodes: nodeValue.map(node => ({
					component_id: parseInt(node.id),
					name: node.data.name,
					component_role:
						node.type === "destination" ? "exporter" : node.type === "source" ? "receiver" : "processor",
					component_name: node.data.component_name,
					config: node.data.config,
					supported_signals: node.data.supported_signals || [],
				})),
				edges: edgeValue.map(edge => ({
					source: edge.source,
					target: edge.target,
				})),
			};
			const syncRes = await pipelineServices.syncPipelineGraph(pipelineId, syncPayload);
			console.log("Sync response:", syncRes);
			setHasDeployError(false);
			localStorage.removeItem("changesLog");
			setIsEditMode(false);
			clearChangesLog();
			toast({
				title: "Success",
				description: "Pipeline successfully deployed",
				duration: 3000,
				variant: "default",
			});
			handleGetPipelineGraph();
		} catch (error) {
			setHasDeployError(true);
			console.error("Error deploying pipeline:", error);
			toast({
				title: "Error",
				description: "Failed to deploy the pipeline",
				duration: 3000,
				variant: "destructive",
			});
		}
	};

	const handleDeletePipeline = async () => {
		try {
			await agentServices.deleteAgentById(pipelineOverviewData?.agent_id);

			await pipelineServices.deletePipelineById(pipelineId);

			setIsOpen(false);
			resetGraph();
			window.location.reload();
		} catch (error) {
			console.error("Error deleting pipeline or collector:", error);
			toast({
				title: "Error",
				description: "Failed to delete pipeline or collector",
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


	const EditForm = async (change: any) => {
		setIsReviewSheetOpen(false)
		setIsEditFormOpen(true)
		setSelectedChange(change)
		const res = await TransporterService.getTransporterForm(change.component_type);
		setForm(res as FormSchema);
		setConfig(change.finalConfig)
	}

	const handleSubmit = () => {
		const log = {
			type: selectedChange.type,
			component_type: selectedChange.component_type,
			id: selectedChange.id,
			name: selectedChange.name,
			status: "edited",
			initialConfig: undefined,
			finalConfig: config,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));
		setIsEditFormOpen(false);
	};

	return (
		<div className="py-4 flex flex-col">
			<div className="flex mb-5 items-center justify-between">
				<div className="flex mb-5 gap-2 items-center">
					<Boxes className="text-gray-700" size={36} />
					<h1 className="text-2xl text-gray-800">{pipelineOverview?.name}</h1>
				</div>
				<div className="flex items-center w-full md:w-auto">
					<div className="flex gap-2 justify-between w-full mb-2">
						<div className="flex gap-2">
							<Sheet
								onOpenChange={open => {
									if(!open){
										setIsEditMode(false);
									}
									if (!open && !hasDeployError) {
										clearChangesLog();
									}
								}}>
								<SheetTrigger asChild>
									<Button className="bg-blue-500">View/Edit Pipeline</Button>
								</SheetTrigger>
								<SheetContent className="w-full sm:max-w-full p-0" side="right">
									<PipelineGraphEditor
										pipelineOverview={pipelineOverview}
										isEditMode={isEditMode}
										setIsEditMode={setIsEditMode}
										changesLog={changesLog}
										handleDeployChanges={handleDeployChanges}
										reactFlowWrapper={reactFlowWrapper}
										nodeValue={nodeValue}
										edgeValue={edgeValue}
										updateNodes={updateNodes}
										updateEdges={updateEdges}
										onConnect={onConnect}
										nodeTypes={nodeTypes}
										setReactFlowInstance={setReactFlowInstance}
										onEdgeClick={onEdgeClick}
										onPaneClick={onPaneClick}
										selectedEdge={selectedEdge}
										edgePopoverPosition={edgePopoverPosition}
										handleDeleteEdge={handleDeleteEdge}
									/>
								</SheetContent>
							</Sheet>
							<DeletePipelineDialog
								isOpen={isOpen}
								setIsOpen={setIsOpen}
								pipelineOverview={pipelineOverview}
								handleDeletePipeline={handleDeletePipeline}
							/>
						</div>
					</div>

				</div>
			</div>
			<div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm p-6">
				<h2 className="text-xl font-semibold text-gray-800 mb-6">Pipeline Overview</h2>

				<div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-6 text-gray-700 text-sm">
					<div>
						<p className="text-gray-500">Name</p>
						<p className="font-medium">{pipelineOverviewData?.name || "-"}</p>
					</div>

					{/* 🔥 STATUS on top right */}
					<div className="flex flex-col">
						<p className="text-gray-500">Status</p>
						<div className="flex items-center gap-2">
							<span
								className={`capitalize px-2 py-1 rounded-full text-xs font-semibold ${
									pipelineOverviewData?.status?.toLowerCase() === "active"
										? "bg-green-100 text-green-700"
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
									className="h-4 w-4 text-gray-500 cursor-pointer hover:text-gray-700 transition-transform hover:rotate-180"
									onClick={handleRefreshStatus}
								/>
							)}
						</div>
					</div>

					<div>
						<p className="text-gray-500">Created At</p>
						<p className="font-medium">{formatTimestampWithDate(pipelineOverviewData?.created_at)}</p>
					</div>

					<div>
						<p className="text-gray-500">Created By</p>
						<p className="font-medium">{pipelineOverviewData?.created_by || "-"}</p>
					</div>

					<div>
						<p className="text-gray-500">Updated At</p>
						<p className="font-medium">{formatTimestampWithDate(pipelineOverviewData?.updated_at)}</p>
					</div>

					<div>
						<p className="text-gray-500">Hostname</p>
						<p className="font-medium">{pipelineOverviewData?.hostname}</p>
					</div>

					<div>
						<p className="text-gray-500">Agent Version</p>
						<p className="font-medium">{pipelineOverviewData?.agent_version}</p>
					</div>

					<div>
						<p className="text-gray-500">IP Address</p>
						<p className="font-medium">{pipelineOverviewData?.ip_address}</p>
					</div>

					<div>
						<p className="text-gray-500">Platform</p>
						<p className="font-medium">{pipelineOverviewData?.platform}</p>

					</div>
				</div>
			</div>}

			{tabs == "overview" && <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
				{!isEditMode && healthMetrics.length > 0 ? (
					healthMetrics.map(metric => (

						<div key={metric.metric_name} className="w-full h-[300px] bg-white rounded-lg shadow-sm p-4">
							<HealthChart
								name={metric.metric_name === "cpu_utilization" ? "CPU Usage" : "Memory Usage"}
								data={metric.data_points.map(point => ({
									timestamp: point.timestamp,
									[metric.metric_name]:
										metric.metric_name === "memory_utilization" ? point.value / (1024 * 1024) : point.value,
								}))}
								y_axis_data_key={metric.metric_name}
								chart_color={getRandomChartColor(metric.metric_name)}
							/>
						</div>
					))
				) : (
					// <div className="col-span-2 text-center py-4 text-gray-500">No health metrics available</div>
					<div className="col-span-2 bg-white rounded-lg shadow-sm p-8 flex flex-col items-center justify-center min-h-[300px]">
						<div className="text-gray-400 mb-2">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								className="h-12 w-12"
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
						<p className="text-gray-500 text-lg font-medium">No Health Metrics Available</p>
						<p className="text-gray-400 text-sm mt-1">
							Health metrics will appear here once data is available
						</p>
					</div>
				)}
			</div>}
		</div>
	);
};

export default ViewPipelineDetails;

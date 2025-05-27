import { Boxes, Edit, RefreshCw, Trash2 } from "lucide-react";
import  { useCallback, useEffect, useMemo, useRef, useState } from "react";
// import EditPipelineYAML from "./EditPipelineYAML";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import {
	Sheet,
	SheetClose,
	SheetContent,
	SheetDescription,
	SheetFooter,
	SheetTitle,
	SheetTrigger,
} from "@/components/ui/sheet";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { useToast } from "@/hooks/use-toast";
import agentServices from "@/services/agentServices";
import pipelineServices from "@/services/pipelineServices";
import { Pipeline } from "@/types/pipeline.types";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import ReactFlow, {
	Background,
	Connection,
	Controls,
	Edge,
	EdgeMouseHandler,
	MiniMap,
	Panel,
	ReactFlowInstance,
} from "reactflow";
import "reactflow/dist/style.css";
import { HealthChart } from "../charts/HealthChart";
import { Label } from "../ui/label";
import { Switch } from "../ui/switch";
import DestinationDropdownOptions from "./DropdownOptions/DestinationDropdownOptions";
import ProcessorDropdownOptions from "./DropdownOptions/ProcessorDropdownOptions";
import SourceDropdownOptions from "./DropdownOptions/SourceDropdownOptions";
import { DestinationNode } from "./Nodes/DestinationNode";
import { ProcessorNode } from "./Nodes/ProcessorNode";
import { SourceNode } from "./Nodes/SourceNode";
import { TransporterService } from "@/services/transporterService";
import { formatTimestampWithDate, getRandomChartColor } from "@/constants";
import { customEnumRenderer } from "./DropdownOptions/CustomEnumControl";
import Yaml from "../YAML/Yaml";



const theme = createTheme({
	components: {
		MuiFormControl: {
			styleOverrides: {
				root: {
					marginBottom: "0.5rem",
				},
			},
		},
	},
});


const renderers = [
	...materialRenderers,
	customEnumRenderer
];
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


	const handleGetPipelineGraph = async () => {
		const res = await pipelineServices.getPipelineGraph(pipelineId);
		const edges = res.edges;
		const VERTICAL_SPACING = 100;

		const updatedNodes = res.nodes.map((node: any, index: number) => {
			const nodeType =
				node.component_role === "receiver"
					? "destination"
					: node.component_role === "exporter"
						? "source"
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
		//filter out edges that have source and target not in updatedNodes id
		// const filteredEdges = updatedEdges.filter(edge =>
		// 	updatedNodes.some(node => node.id === edge.source && updatedNodes.some(node => node.id === edge.target)),
		// );
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

	// const handleDeleteEdge = useCallback(() => {
	// 	if (selectedEdge) {
	// 		deleteEdge(selectedEdge);
	// 		setSelectedEdge(null);
	// 	}
	// }, [selectedEdge, deleteEdge]);

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
						node.type === "destination" ? "receiver" : node.type === "source" ? "exporter" : "processor",
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

			await pipelineServices.deletePipelineById(pipelineId);
			setIsOpen(false);
			resetGraph();
			window.location.reload();
		} catch (error) {
			console.error("Error deleting pipeline or agents:", error);
			toast({
				title: "Error",
				description: "Failed to delete pipeline or agents",
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
		<div className="flex flex-col h-[100vh] overflow-hidden">
			{/* Header */}
			<div className="flex items-center justify-between px-6 py-4 border-b bg-white flex-shrink-0">
				<div className="flex gap-2 items-center">
					<Boxes className="text-gray-700" size={32} />
					<h1 className="text-xl text-gray-800 font-semibold">{pipelineOverview?.name}</h1>
				</div>
				<div className="flex items-center w-full md:w-auto">
					<div className="flex gap-2 justify-between w-full mb-2">
						<div className="flex gap-2">
							<Sheet
								onOpenChange={open => {
									if (!open) {
										setIsEditMode(false);
									}
									if (!open && !hasDeployError) {
										clearChangesLog();
									}
								}}
							>
								<SheetTrigger asChild>
									<Button className="bg-blue-500">View/Edit Pipeline</Button>
								</SheetTrigger>
								<SheetContent className="w-full sm:max-w-full p-0" side="right">
									<div className="flex justify-between items-center p-4 border-b">
										<div className="flex items-center space-x-2">
											<div className="text-xl font-medium">{pipelineOverview?.name}</div>
										</div>
										<div className="flex items-center mr-6">
											<div className="mx-4 flex items-center space-x-2">
												<Switch id="edit-mode" checked={isEditMode} onCheckedChange={setIsEditMode} />
												<Label htmlFor="edit-mode">Edit Mode</Label>
											</div>
											<Sheet open={isReviewSheetOpen || isEditFormOpen} onOpenChange={open => {
												setIsReviewSheetOpen(open && !isEditFormOpen);
												setIsEditFormOpen(open && isEditFormOpen);
											}}>
												<SheetTrigger asChild>
													<Button className="rounded-md px-6" disabled={!isEditMode}>
														Review
													</Button>
												</SheetTrigger>
												<SheetContent className="w-[30rem]">
													{isReviewSheetOpen && (
														<div>
															<SheetTitle>Pending Changes</SheetTitle>
															<SheetDescription>
																<div className="flex flex-col gap-6 mt-4 overflow-auto h-[40rem]">
																	{changesLog.map((change, index) => (
																		<div key={index} className="flex justify-between items-center">
																			<div className="flex flex-col">
																				<p className="text-lg">{change.type}</p>
																				<p className="text-lg text-gray-800">{change.name}</p>
																			</div>
																			<div className="flex justify-end gap-3 items-center">
																				<p
																					className={`${change.status == "added"
																						? "text-green-500"
																						: change.status == "deleted"
																							? "text-red-500"
																							: "text-gray-600"
																						} text-lg`}
																				>
																					[{change.status}]
																				</p>
																				<Edit onClick={() => EditForm(change)} className="w-6 h-6 cursor-pointer" />
																			</div>
																		</div>
																	))}
																</div>
															</SheetDescription>
															<SheetClose className="flex justify-end mt-4 w-full">
																<div>

																	<Button onClick={handleDeployChanges} className="bg-blue-500">
																		Deploy Changes
																	</Button>
																</div>
															</SheetClose>
														</div>
													)}
													{isEditFormOpen && selectedChange && (
														<div className="flex flex-col gap-4 p-4">
															<div className="flex gap-3 items-center">
																<p className="text-lg bg-gray-500 items-center rounded-lg p-2 px-3 m-1 text-white">→|</p>
																<h2 className="text-xl font-bold">{selectedChange.name}</h2>
															</div>
															<p className="text-gray-500">
																Generate the defined log type at the rate desired{" "}
																<span className="text-blue-500 underline">Documentation</span>
															</p>
															<ThemeProvider theme={theme}>
																<div className="">
																	<div className="p-3  ">
																		<div className="overflow-y-auto h-[32rem] pt-3">
																			{isEditFormOpen && form && <JsonForms
																				data={config}
																				schema={form}
																				renderers={renderers}
																				cells={materialCells}
																				onChange={({ data }) => {
																					setConfig(data);
																				}}
																			/>}
																		</div>
																	</div>
																</div>
															</ThemeProvider>
															<SheetFooter>
																<SheetClose>
																	<div className="flex gap-3">
																		<Button className="bg-blue-500" onClick={handleSubmit}>
																			Add Source
																		</Button>
																		<Button variant={"outline"} onClick={() => setIsEditFormOpen(false)}>
																			Discard Changes
																		</Button>
																	</div>
																</SheetClose>
															</SheetFooter>
														</div>
													)}
												</SheetContent>
											</Sheet>
										</div>
									</div>
									<div style={{ height: "77vh", backgroundColor: "#f9f9f9" }}>
										<ReactFlow
											nodes={nodeValue}
											edges={edgeValue}
											onNodesChange={updateNodes}
											onEdgesChange={updateEdges}
											onConnect={isEditMode ? onConnect : undefined}
											nodeTypes={nodeTypes}
											onInit={setReactFlowInstance}
											onEdgeClick={onEdgeClick}
											onPaneClick={onPaneClick}
											nodesDraggable={isEditMode}
											nodesConnectable={isEditMode}
											elementsSelectable={isEditMode}
											onlyRenderVisibleElements
											proOptions={{ hideAttribution: true }}
											fitView
										>
											<Background />
											<Controls />
											<MiniMap />
											{selectedEdge && isEditMode && (
												<Panel
													position="top-left"
													style={{
														position: "absolute",
														left: edgePopoverPosition.x,
														top: edgePopoverPosition.y,
														transform: "translate(-50%, -50%)",
														background: "white",
														padding: "8px",
														borderRadius: "4px",
														boxShadow: "0 2px 4px rgba(0,0,0,0.2)",
														zIndex: 10,
													}}
												>
													<Trash2 onClick={handleDeleteEdge} className="text-red-500 cursor-pointer" size={16} />
												</Panel>
											)}
										</ReactFlow>
									</div>
									<div className="bg-gray-100 h-1/5 p-4 rounded-lg">
										<div className="flex justify-around gap-2">
											<div className="flex items-center">
												<SourceDropdownOptions disabled={!isEditMode} />
											</div>
											<div className="flex items-center">
												<ProcessorDropdownOptions disabled={!isEditMode} />
											</div>

											<div className="flex items-center">
												<DestinationDropdownOptions disabled={!isEditMode} />
											</div>
										</div>
									</div>
								</SheetContent>
							</Sheet>
							<Dialog open={isOpen} onOpenChange={setIsOpen}>
								<DialogTrigger asChild>
									<Button variant="destructive">Delete Pipeline</Button>
								</DialogTrigger>
								<DialogContent className="sm:max-w-[28rem] h-[14rem]">
									<DialogHeader>
										<DialogTitle className="text-red-500 text-xl">Delete Pipeline</DialogTitle>
										<DialogDescription className="text-md text-gray-700">
											Are you sure you want to delete this Pipeline?
										</DialogDescription>
									</DialogHeader>
									<div className="flex flex-col">
										<p className="text-gray-600">Pipeline Id: {pipelineOverview?.id} </p>
										<p className="text-gray-600">Pipeline Name: {pipelineOverview?.name}</p>
									</div>
									<DialogFooter>
										<DialogClose asChild>
											<Button>Cancel</Button>
										</DialogClose>
										<Button onClick={handleDeletePipeline} variant={"destructive"}>
											Delete
										</Button>
									</DialogFooter>
								</DialogContent>
							</Dialog>
						</div>
					</div>

				</div>
			</div>
			<div>
				<ul className="flex border-b">
					<li
						className={`mr-6 cursor-pointer py-2 ${tabs === "overview" ? "border-b-2 border-blue-500" : ""}`}
						onClick={() => setTabs("overview")}
					>
						Overview
					</li>
					<li
						className={`mr-6 cursor-pointer py-2 ${tabs === "yaml" ? "border-b-2 border-blue-500" : ""}`}
						onClick={() => setTabs("yaml")}
					>
						YAML
					</li>
				</ul>
			</div>
			{/* Main Content */}
			<div className="flex-1 overflow-auto mt-4">
				{tabs == "overview" && (
					<>
						<div className="w-full bg-white rounded-lg border border-gray-200 shadow-sm p-6 mb-4">
							<h2 className="text-xl font-semibold text-gray-800 mb-6">Pipeline Overview</h2>
							<div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-6 text-gray-700 text-sm">
								<div>
									<p className="text-gray-500">Name</p>
									<p className="font-medium">{pipelineOverviewData?.name || "-"}</p>
								</div>
								<div className="flex flex-col">
									<p className="text-gray-500">Status</p>
									<div className="flex items-center gap-2">
										<span
											className={`capitalize px-2 py-1 rounded-full text-xs font-semibold ${pipelineOverviewData?.status?.toLowerCase() === "connected"
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
						</div>
						<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
							{!isEditMode && healthMetrics.length > 0 ? (
								healthMetrics.map(metric => (
									<div key={metric.metric_name} className="w-full h-[300px] bg-white rounded-lg shadow-sm ">
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
								<div className="col-span-2 bg-white rounded-lg shadow-sm flex flex-col items-center justify-center min-h-[300px]">
									<div className="text-gray-400 mb-2">
										<svg
											xmlns="http://www.w3.org/2000/svg"
											className="h-12 w-12"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
										>
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
						</div>
					</>
				)}
				{tabs == "yaml" && <Yaml jsonforms={pipelineOverviewData?.config} />}
			</div>
		</div>
	);
};

export default ViewPipelineDetails;

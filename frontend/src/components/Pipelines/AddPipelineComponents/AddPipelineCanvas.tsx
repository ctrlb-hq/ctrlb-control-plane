import React, { useState, useCallback, useRef } from "react";
import ReactFlow, {
	MiniMap,
	Controls,
	Background,
	// addEdge,
	// useEdgesState,
	// Edge,
	// Connection,
	ReactFlowInstance,
	Connection,
	EdgeMouseHandler,
	Panel,
	Edge,
} from "reactflow";

import {
	Sheet,
	SheetClose,
	SheetContent,
	SheetDescription,
	SheetHeader,
	SheetTitle,
	SheetTrigger,
} from "@/components/ui/sheet";

import "reactflow/dist/style.css";
import { SourceNode } from "../Nodes/SourceNode";
import { ProcessorNode } from "../Nodes/ProcessorNode";
import { DestinationNode } from "../Nodes/DestinationNode";
import SourceDropdownOptions from "../DropdownOptions/SourceDropdownOptions";
import ProcessorDropdownOptions from "../DropdownOptions/ProcessorDropdownOptions";
import DestinationDropdownOptions from "../DropdownOptions/DestinationDropdownOptions";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { Button } from "../../ui/button";
import { Edit, Trash2 } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import pipelineServices from "@/services/pipelineServices";
import { useToast } from "@/hooks/use-toast";

// Node types mapping
const nodeTypes = {
	source: SourceNode,
	processor: ProcessorNode,
	destination: DestinationNode,
};

const AddPipelineCanvas = () => {
	const {
		nodeValue,
		edgeValue,
		updateNodes,
		updateEdges,
		connectNodes,
		deleteEdge,
	} = useGraphFlow();
	const reactFlowWrapper = useRef<HTMLDivElement>(null);
	const [_reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
	const [isEditMode, setIsEditMode] = useState(true);
	const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });
	const { changesLog } = usePipelineChangesLog();
	const { toast } = useToast();
	const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);

	const pipelineName = localStorage.getItem("pipelinename");
	const createdBy = localStorage.getItem("userEmail");
	const agentIds = JSON.parse(localStorage.getItem("selectedAgentIds") || "");

	const onConnect = useCallback(
		(params: Edge | Connection) => {
			connectNodes(params);
		},
		[connectNodes],
	);

	const handleDeleteEdge = useCallback(() => {
		if (selectedEdge) {
			deleteEdge(selectedEdge);
			setSelectedEdge(null);
		}
	}, [selectedEdge, deleteEdge]);

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

	// Close popover when clicking elsewhere
	const onPaneClick = useCallback(() => {
		setSelectedEdge(null);
	}, []);

	const onDragOver = useCallback((event: React.DragEvent) => {
		event.preventDefault();
		event.dataTransfer.dropEffect = "move";
	}, []);

	// const onDrop = useCallback(
	// 	(event: React.DragEvent) => {
	// 		event.preventDefault();
	// 		const type = event.dataTransfer.getData("application/nodeType");
	// 		if (!type || !reactFlowInstance) return;
	// 		const position = reactFlowInstance.project({ x: event.clientX, y: event.clientY });
	// 		let nodeData;
	// 		const id = `node_${Date.now()}`;

	// 		updateNodes([{ item: { id, type, data: nodeData, position }, type: 'add' }]);
	// 	},
	// 	[reactFlowInstance, nodeValue, updateNodes],
	// );

	const addPipeline = async () => {
		const pipelinePayload = {
			name: pipelineName,
			created_by: createdBy,
			agent_ids: [parseInt(agentIds)],
			pipeline_graph: {
				nodes: nodeValue.map(node => ({
					component_id: parseInt(node.id),
					name: node.data.name,
					component_role:
						node.type === "source" ? "exporter" : node.type === "destination" ? "receiver" : "processor",
					component_name: node.data.component_name,
					supported_signals: node.data.supported_signals,
					config: node.data.config,
				})),
				edges: edgeValue.map(edge => ({
					source: edge.source,
					target: edge.target,
				})),
			},
		};
		console.log("payload", pipelinePayload);
		const res = await pipelineServices.addPipeline(pipelinePayload);
		console.log(res);
	};

	const handleDeployChanges = () => {
		try {
			addPipeline();
			localStorage.removeItem("Sources");
			localStorage.removeItem("Destination");
			localStorage.removeItem("pipelinename");
			localStorage.removeItem("selectedAgentIds");
			localStorage.removeItem("Nodes");
			localStorage.removeItem("changesLog");
			localStorage.removeItem("PipelineEdges");
			setTimeout(() => {
				toast({
					title: "Success",
					description: "Successfully deployed the pipeline",
					duration: 3000,
				});
				window.location.reload();
			}, 2000);
		} catch (error) {
			console.error("Error deploying pipeline:", error);
			toast({
				title: "Error",
				description: "Failed to add and deploy the pipeline",
				duration: 3000,
				variant: "destructive",
			});
		}
	};

	return (
		<>
			<SheetContent>
				<SheetHeader>
					<div className="flex justify-between items-center p-2 border-b">
						<SheetTitle>
							<div className="flex items-center space-x-2">
								<div className="text-xl font-medium">{pipelineName}</div>
							</div>
						</SheetTitle>

						<div className="flex items-center mx-4">
							<Sheet>
								<SheetTrigger asChild>
									<Button className="rounded-full px-6">Review</Button>
								</SheetTrigger>
								<SheetContent className="w-[30rem]">
									<SheetTitle>Pending Changes</SheetTitle>
									<SheetDescription>
										<div className="flex flex-col gap-6 mt-4 overflow-auto h-[40rem]">
											{changesLog.map((change, index) => (
												<div key={index} className="flex justify-between items-center">
													<div className="flex flex-col">
														{/* <p className="text-lg capitalize">{change.component_role}</p> */}
														<p className="text-lg text-gray-800 capitalize">{change.name}</p>
													</div>
													<div className="flex justify-end gap-3 items-center">
														<p
															className={`${change.status == "edited" ? "text-gray-500" : change.status == "deleted" ? "text-red-500" : "text-green-600"} text-lg`}
														>
															[{change.status ? change.status : "Added"}]
														</p>
														<Edit size={20} />
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
								</SheetContent>
							</Sheet>
							<div className="mx-4 flex items-center space-x-2">
								<Switch id="edit-mode" checked={isEditMode} onCheckedChange={setIsEditMode} />
								<Label htmlFor="edit-mode">Edit Mode</Label>
							</div>
						</div>
					</div>
				</SheetHeader>
				<div className="w-full flex flex-col gap-2 h-screen p-4">
					<div className="h-4/5 border-2 border-gray-200 rounded-lg">
						<ReactFlow
							nodes={nodeValue}
							edges={edgeValue}
							onNodesChange={updateNodes}
							onEdgesChange={updateEdges}
							nodesConnectable={isEditMode}
							nodesDraggable={isEditMode}
							elementsSelectable={isEditMode}
							onConnect={onConnect}
							onInit={setReactFlowInstance}
							onEdgeClick={onEdgeClick}
							onPaneClick={onPaneClick}
							// onDrop={onDrop}
							onDragOver={onDragOver}
							nodeTypes={nodeTypes}
							fitView
						>
							<MiniMap />
							<Controls />
							<Background color="#aaa" gap={16} />
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
					<div className=" p-2 pb-4">
						<div className="flex justify-center gap-6 items-center">
							<div className="flex gap-6 bg-gray-100 p-4 rounded-lg">
								<SourceDropdownOptions disabled={!isEditMode} />
								<ProcessorDropdownOptions disabled={!isEditMode} />
								<DestinationDropdownOptions disabled={!isEditMode} />
							</div>
						</div>
					</div>
				</div>
			</SheetContent>
		</>
	);
};

export default AddPipelineCanvas;

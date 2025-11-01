import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
	Sheet,
	SheetClose,
	SheetContent,
	SheetDescription,
	SheetTitle,
	SheetTrigger,
} from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { useToast } from "@/hooks/useToast";
import pipelineServices from "@/services/pipeline";
import { ComponentService } from "@/services/component";
import { Edit, Trash2, Loader2 } from "lucide-react";
import { Dispatch, SetStateAction, useCallback, useEffect, useMemo, useRef, useState } from "react";
import ReactFlow, {
	Background,
	Connection,
	Controls,
	Edge,
	EdgeMouseHandler,
	MiniMap,
	NodeProps,
	Panel,
	ReactFlowInstance,
} from "reactflow";
import GenericNode from "@/components/pipelines/editor/GenericNode";
import PluginDropdownOptions from "@/components/pipelines/editor/PluginDropdownOptions";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";

const PipelineEditorSheet = ({
	pipelineId,
	name,
	setIsSheetOpen,
	isEditModeStart,
}: {
	pipelineId: string;
	name: string;
	setIsSheetOpen: Dispatch<SetStateAction<boolean>>;
	isEditModeStart: boolean;
}) => {
	const [isEditMode, setIsEditMode] = useState(isEditModeStart);
	const [isReviewSheetOpen, setIsReviewSheetOpen] = useState(false);
	const [isEditFormOpen, setIsEditFormOpen] = useState(false);
	const [form, setForm] = useState<any>({});
	const [config, setConfig] = useState<object>({});
	const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({
		type: "VerticalLayout",
		elements: [],
	});
	const [selectedChange, setSelectedChange] = useState<any>(null);
	const [_hasDeployError, setHasDeployError] = useState(false);
	const {
		nodeValue,
		edgeValue,
		updateNodes,
		updateEdges,
		setEdgeValueDirect,
		setNodeValueDirect,
		connectNodes,
		changesLog,
		deleteEdge,
		clearChangesLog,
		updateNodeConfig
	} = useGraphFlow();
	const reactFlowWrapper = useRef<HTMLDivElement>(null);
	const [_reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
	const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
	const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });
	const { toast } = useToast();

	const nodeTypes = useMemo(
		() => ({
			source: (props: NodeProps) => <GenericNode {...props} type="source" />,
			processor: (props: NodeProps) => <GenericNode {...props} type="processor" />,
			destination: (props: NodeProps) => <GenericNode {...props} type="destination" />,
		}),
		[],
	);

	const fetchGraph = async () => {
		setNodeValueDirect([]);
		setEdgeValueDirect([]);
		clearChangesLog();

		const res = await pipelineServices.getPipelineGraph(pipelineId);
		const VERTICAL_SPACING = 100;

		const updatedNodes = res.nodes.map((node: any, index: number) => {
			const nodeType =
				node.component_role === "receiver"
					? "source"
					: node.component_role === "exporter"
						? "destination"
						: "processor";
			const x = nodeType === "source" ? 50 : nodeType === "destination" ? 400 : 225;
			const y = 100 + index * VERTICAL_SPACING;
			return {
				id: node.component_id.toString(),
				type: nodeType,
				position: { x, y },
				data: node,
			};
		});
		const updatedEdges = res.edges.map((edge: any) => ({
			id: `edge-${edge.source}-${edge.target}`,
			source: edge.source,
			target: edge.target,
			animated: true,
		}));
		setNodeValueDirect(updatedNodes);
		setEdgeValueDirect(updatedEdges);
	};

	const onConnect = useCallback(
		(params: Edge | Connection) => {
			connectNodes(params);
		},
		[connectNodes],
	);

	const onEdgeClick: EdgeMouseHandler = useCallback(
		(event, edge) => {
			if (!isEditMode) return;
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
		if (!selectedEdge) return;
		deleteEdge(selectedEdge);
		setSelectedEdge(null);
	}, [selectedEdge, deleteEdge]);

	const handleDeployChanges = async () => {
		const toastId = toast({
			title: "Deploying changes...",
			description: (
				<div className="flex items-center">
					<Loader2 className="mr-2 h-4 w-4 animate-spin" />
					Please wait while we sync the pipeline
				</div>
			),
			variant: "default",
			duration: Infinity,
			className: "bg-blue-100 text-blue-800 border border-blue-200",
		});

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

			await pipelineServices.syncPipelineGraph(pipelineId, syncPayload);

			toastId.dismiss();

			toast({
				title: "Success",
				description: "Changes deployed successfully",
				variant: "default",
			});

			setHasDeployError(false);
			setIsEditMode(false);
			clearChangesLog();
			fetchGraph();
			setIsSheetOpen(false);
		} catch (err) {
			console.error("Deploy error:", err);
			toastId.dismiss();

			setHasDeployError(true);
			toast({
				title: "Error",
				description: "Failed to deploy changes",
				variant: "destructive",
			});
		}
	};

	const EditForm = async (change: any) => {
		setIsReviewSheetOpen(false);
		setIsEditFormOpen(true);
		setSelectedChange(change);
		console.log(changesLog);
		console.log(change);
		const schema = await ComponentService.getTransporterForm(change.component_type);
		const ui = await ComponentService.getTransporterUiSchema(change.component_type);
		setForm(schema);
		setUiSchema(ui);
		setConfig(change.finalConfig);
	};

	const handleSubmit = useCallback((submittedConfig: any) => {
		if (selectedChange) {
			updateNodeConfig(selectedChange.id, submittedConfig);
			setNodeValueDirect((nodes) =>
				nodes.map((node) =>
				node.id === selectedChange.id
					? {
						...node,
						data: {
						...node.data,
						config: submittedConfig,
						},
					}
					: node
				)
			);
			setSelectedChange((prev) =>
				prev ? { ...prev, finalConfig: submittedConfig } : prev
			);
		}
		setIsEditFormOpen(false);
	},
		[selectedChange, setNodeValueDirect, updateNodeConfig]
	);

	const onPaneClick = useCallback(() => {
		setSelectedEdge(null);
	}, []);

	useEffect(() => {
		fetchGraph();
	}, [pipelineId]);

	return (
		<>
			<div className="flex justify-between items-center p-4 border-b">
				<div className="text-xl font-medium">{name}</div>
				<div className="flex items-center mr-6 space-x-4">
					<Switch id="edit-mode" checked={isEditMode} onCheckedChange={setIsEditMode} />
					<Label htmlFor="edit-mode">Edit Mode</Label>
					<Sheet
						open={isReviewSheetOpen || isEditFormOpen}
						onOpenChange={open => {
							setIsReviewSheetOpen(open && !isEditFormOpen);
							setIsEditFormOpen(open && isEditFormOpen);
						}}>
						<SheetTrigger asChild>
							<Button disabled={!isEditMode}>Review</Button>
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
														<p className="text-gray-800">{change.name}</p>
													</div>
													<div className="flex items-center gap-3">
														<p
															className={`text-lg ${change.status === "deleted" ? "text-red-500" : change.status === "added" ? "text-green-500" : "text-gray-500"}`}>
															[{change.status}]
														</p>
														{change.type !== "Edge" && (
															<Edit onClick={() => EditForm(change)} className="w-6 h-6 cursor-pointer" />
														)}
													</div>
												</div>
											))}
										</div>
									</SheetDescription>
									<SheetClose className="mt-4">
										<Button onClick={handleDeployChanges} className="bg-blue-500">
											Deploy Changes
										</Button>
									</SheetClose>
								</div>
							)}
							{isEditFormOpen && selectedChange && (
								<NodeSidePanel
									title={selectedChange.name}
									formSchema={form}
									uiSchema={uiSchema}
									config={config}
									setConfig={setConfig}
									submitLabel="Apply"
									onSubmit={handleSubmit}
									onDiscard={() => setSelectedChange(null)}
									showDelete={false}
								/>
							)}
						</SheetContent>
					</Sheet>
				</div>
			</div>
			<div
				ref={reactFlowWrapper}
				style={{ height: "92.5vh", width: "100vw", backgroundColor: "#f9f9f9" }}>
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
					fitView>
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
							}}>
							<Trash2 onClick={handleDeleteEdge} className="text-red-500 cursor-pointer" size={16} />
						</Panel>
					)}
				</ReactFlow>
				<div
					style={{
						position: "absolute",
						bottom: "5rem", // distance from bottom
						left: "50%",
						transform: "translateX(-50%)",
						backgroundColor: "#f1f5f9",
						padding: "12px 24px",
						borderRadius: "8px",
						boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
						display: "flex",
						gap: "12px",
						zIndex: 20,
					}}>
					<PluginDropdownOptions
						kind="receiver"
						nodeType="source"
						label="Source"
						dataType="receiver"
						disabled={!isEditMode}
					/>
					<PluginDropdownOptions
						kind="processor"
						nodeType="processor"
						label="Processor"
						dataType="receiver"
						disabled={!isEditMode}
					/>
					<PluginDropdownOptions
						kind="exporter"
						nodeType="destination"
						label="Destination"
						dataType="exporter"
						disabled={!isEditMode}
					/>
				</div>
			</div>
		</>
	);
};

export default PipelineEditorSheet;

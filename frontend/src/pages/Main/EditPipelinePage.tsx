import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams, useLocation } from "react-router-dom";
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
import { Edit, Loader2, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import pipelineServices from "@/services/pipeline";
import GenericNode from "@/components/pipelines/editor/GenericNode";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";
import { ComponentService } from "@/services/component";
import PluginDropdownOptions from "@/components/pipelines/editor/PluginDropdownOptions";


type AgentType = "otel" | "fluent-bit";

const EditPipelinePage = () => {
  const { pipelineId } = useParams<{ pipelineId: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  const pipelineName = (location.state as any)?.pipelineName ?? "Pipeline Editor";

  const [isEditMode, setIsEditMode] = useState<boolean | false>(false);
  const [agentType, setAgentType] = useState<AgentType>("otel");

  // Load agent type from localStorage on mount
  useEffect(() => {
    const storedAgentType = localStorage.getItem("agentType") as AgentType;
    if (storedAgentType && (storedAgentType === "otel" || storedAgentType === "fluent-bit")) {
      setAgentType(storedAgentType);
    }
  }, []);
  const [isReviewSheetOpen, setIsReviewSheetOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [form, setForm] = useState<any>({});
  const [config, setConfig] = useState<object>({});
  const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({
    type: "VerticalLayout",
    elements: [],
  });
  const [selectedChange, setSelectedChange] = useState<any>(null);
  const [isDeploying, setIsDeploying] = useState(false);

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
    updateNodeConfig,
  } = useGraphFlow();

  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [_reactFlowInstance, setReactFlowInstance] =
    useState<ReactFlowInstance | null>(null);
  const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
  const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });

  const { showSnackbar } = useGlobalSnackbar();

  const nodeTypes = useMemo(
    () => ({
      source: (props: NodeProps) => (
        <GenericNode {...props} type="source" isEditMode={isEditMode} />
      ),
      processor: (props: NodeProps) => (
        <GenericNode {...props} type="processor" isEditMode={isEditMode} />
      ),
      destination: (props: NodeProps) => (
        <GenericNode {...props} type="destination" isEditMode={isEditMode} />
      ),
    }),
    [isEditMode]
  );

  const fetchGraph = async () => {
    if (!pipelineId) return;

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

  useEffect(() => {
    fetchGraph();
  }, [pipelineId]);

  const onConnect = useCallback(
    (params: Edge | Connection) => connectNodes(params),
    [connectNodes]
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
    [isEditMode]
  );

  const handleDeleteEdge = useCallback(() => {
    if (!selectedEdge) return;
    deleteEdge(selectedEdge);
    setSelectedEdge(null);
  }, [selectedEdge, deleteEdge]);

  const handleDeployChanges = async () => {
    try {
      setIsDeploying(true);
      showSnackbar("Deploying changes…", "loading");

      const syncPayload = {
        nodes: nodeValue.map((node) => ({
          component_id: parseInt(node.id),
          name: node.data.name,
          component_role:
            node.type === "destination"
              ? "exporter"
              : node.type === "source"
              ? "receiver"
              : "processor",
          component_name: node.data.component_name,
          config: node.data.config,
          supported_signals: node.data.supported_signals || [],
        })),
        edges: edgeValue.map((edge) => ({
          source: edge.source,
          target: edge.target,
        })),
      };

      await pipelineServices.syncPipelineGraph(pipelineId!, syncPayload);

      showSnackbar("Changes deployed successfully", "success");
      clearChangesLog();
      setIsEditMode(false);

      navigate("/home");
    } catch (err) {
      console.error(err);
      showSnackbar("Failed to deploy changes", "error");
    } finally {
      setIsDeploying(false);
    }
  };

  const EditForm = async (change: any) => {
    setIsReviewSheetOpen(false);
    setIsEditFormOpen(true);
    setSelectedChange(change);

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

    const handleCloseEditor = useCallback(() => {
      setNodeValueDirect([]);
      setEdgeValueDirect([]);
      clearChangesLog();

      setIsEditMode(false);
      setIsReviewSheetOpen(false);
      setIsEditFormOpen(false);
      setSelectedEdge(null);

      navigate("/home");
    }, [
      clearChangesLog,
      navigate,
      setEdgeValueDirect,
      setNodeValueDirect,
    ]);

  return (
    <>
      <div className="flex justify-between items-center p-4 border-b">
        <div className="text-xl font-medium">{pipelineName}</div>
        <div className="flex items-center gap-4">
          <Switch checked={isEditMode} onCheckedChange={setIsEditMode} />
          <Label>Edit Mode</Label>
          <Sheet
            open={isReviewSheetOpen || isEditFormOpen}
            onOpenChange={(open) => {
              setIsReviewSheetOpen(open && !isEditFormOpen);
              setIsEditFormOpen(open && isEditFormOpen);
            }}
          >
            <div className="flex items-center gap-2">
              <SheetTrigger asChild>
                <Button disabled={!isEditMode}>Review</Button>
              </SheetTrigger>

              <Button
                variant="outline"
                onClick={handleCloseEditor}
              >
                Close
              </Button>
            </div>
            <SheetContent className="w-[30rem]">
              {isReviewSheetOpen && (
                <>
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
                  <div className="mt-4">
										<Button
											onClick={handleDeployChanges}
											className="bg-blue-500 flex items-center gap-2"
											disabled={isDeploying}
										>
											{isDeploying ? (
											<>
												<Loader2 className="h-4 w-4 animate-spin" />
												Deploying…
											</>
											) : (
											"Deploy Changes"
											)}
										</Button>
									</div>
                </>
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
						kind={agentType === "fluent-bit" ? "input" : "receiver"}
						nodeType="source"
						label="Source"
						dataType={agentType === "fluent-bit" ? "input" : "receiver"}
						disabled={!isEditMode}
						agentType={agentType}
					/>
					<PluginDropdownOptions
						kind={agentType === "fluent-bit" ? "filter" : "processor"}
						nodeType="processor"
						label="Processor"
						dataType={agentType === "fluent-bit" ? "filter" : "processor"}
						disabled={!isEditMode}
						agentType={agentType}
					/>
					<PluginDropdownOptions
						kind={agentType === "fluent-bit" ? "output" : "exporter"}
						nodeType="destination"
						label="Destination"
						dataType={agentType === "fluent-bit" ? "output" : "exporter"}
						disabled={!isEditMode}
						agentType={agentType}
					/>
				</div>
			</div>
    </>
  );
};

export default EditPipelinePage;

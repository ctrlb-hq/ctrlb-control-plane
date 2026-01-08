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
import { Trash2 } from "lucide-react";
import {FormControlLabel, Button, Switch} from "@mui/material";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import pipelineServices from "@/services/pipeline";
import GenericNode from "@/components/pipelines/editor/GenericNode";
import PluginDropdownOptions from "@/components/pipelines/editor/PluginDropdownOptions";
import ReviewDrawer from "@/components/pipelines/editor/ReviewDrawer";

type AgentType = "otel" | "fluent-bit";
type FlowNodeType = "source" | "processor" | "destination";
type AgentRole =
  | "input"
  | "filter"
  | "output"
  | "receiver"
  | "processor"
  | "exporter";

const FLOW_TO_AGENT_ROLE: Record<AgentType, Record<FlowNodeType, string>> = {
  "fluent-bit": {
    source: "input",
    processor: "filter",
    destination: "output",
  },
  otel: {
    source: "receiver",
    processor: "processor",
    destination: "exporter",
  },
};

const AGENT_ROLE_TO_FLOW_TYPE: Record<AgentRole, FlowNodeType> = {
  input: "source",
  filter: "processor",
  output: "destination",
  receiver: "source",
  processor: "processor",
  exporter: "destination",
};

const EditPipelinePage = () => {
  const { pipelineId } = useParams<{ pipelineId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const pipelineName =
    (location.state as any)?.pipelineName ?? "Pipeline Editor";
  const [isEditMode, setIsEditMode] = useState(false);
  const [agentType, setAgentType] = useState<AgentType>("otel");
  const [isReviewOpen, setIsReviewOpen] = useState(false);
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
  const [_rfInstance, setRFInstance] =
    useState<ReactFlowInstance | null>(null);

  const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
  const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });

  const { showSnackbar } = useGlobalSnackbar();

  useEffect(() => {
    const stored = localStorage.getItem("agentType") as AgentType;
    if (stored === "otel" || stored === "fluent-bit") {
      setAgentType(stored);
    }
  }, []);

  const nodeTypes = useMemo(
    () => ({
      source: (props: NodeProps) => (
        <GenericNode
          {...props}
          type={agentType === "fluent-bit" ? "input" : "receiver"}
          isEditMode={isEditMode}
        />
      ),
      processor: (props: NodeProps) => (
        <GenericNode
          {...props}
          type={agentType === "fluent-bit" ? "filter" : "processor"}
          isEditMode={isEditMode}
        />
      ),
      destination: (props: NodeProps) => (
        <GenericNode
          {...props}
          type={agentType === "fluent-bit" ? "output" : "exporter"}
          isEditMode={isEditMode}
        />
      ),
    }),
    [agentType, isEditMode],
  );

  const fetchGraph = useCallback(async () => {
    if (!pipelineId) return;

    setNodeValueDirect([]);
    setEdgeValueDirect([]);
    clearChangesLog();

    const res = await pipelineServices.getPipelineGraph(pipelineId);

    const updatedNodes = res.nodes
      .map((node: any, index: number) => {
        const flowType =
          AGENT_ROLE_TO_FLOW_TYPE[node.component_role as AgentRole];
        if (!flowType) return null;

        return {
          id: String(node.component_id),
          type: flowType,
          position: {
            x: flowType === "source" ? 50 : flowType === "processor" ? 225 : 400,
            y: 100 + index * 100,
          },
          data: node,
        };
      })
      .filter(Boolean);

    const updatedEdges = res.edges.map((edge: any) => ({
      id: `edge-${edge.source}-${edge.target}`,
      source: String(edge.source),
      target: String(edge.target),
      animated: true,
    }));

    setNodeValueDirect(updatedNodes);
    setEdgeValueDirect(updatedEdges);
  }, [pipelineId]);

  useEffect(() => {
    fetchGraph();
  }, [fetchGraph]);

  const onConnect = useCallback(
    (params: Edge | Connection) => connectNodes(params),
    [connectNodes],
  );

  const onEdgeClick: EdgeMouseHandler = useCallback(
    (event, edge) => {
      if (!isEditMode) return;
      const rect = reactFlowWrapper.current?.getBoundingClientRect();
      if (!rect) return;
      setEdgePopoverPosition({
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      });
      setSelectedEdge(edge);
    },
    [isEditMode],
  );

  const handleDeployChanges = async () => {
    try {
      setIsDeploying(true);
      showSnackbar("Deploying changes…", "loading");

      const roleMap = FLOW_TO_AGENT_ROLE[agentType];

      await pipelineServices.syncPipelineGraph(pipelineId!, {
        nodes: nodeValue.map((node) => ({
          component_id: Number(node.id),
          name: node.data.name,
          component_role: roleMap[node.type as FlowNodeType],
          component_name: node.data.component_name,
          config: node.data.config,
          supported_signals: node.data.supported_signals ?? [],
        })),
        edges: edgeValue.map((e) => ({
          source: e.source,
          target: e.target,
        })),
      });

      showSnackbar("Changes deployed successfully", "success");
      clearChangesLog();
      navigate("/home");
    } catch {
      showSnackbar("Failed to deploy changes", "error");
    } finally {
      setIsDeploying(false);
    }
  };

  return (
    <>
      <div className="flex justify-between items-center p-4 border-b">
        <div className="text-xl font-medium">{pipelineName}</div>

        <div className="flex items-center gap-4">
          <FormControlLabel
            control={
              <Switch
                checked={isEditMode}
                onChange={(_, v) => setIsEditMode(v)}
              />
            }
            label="Edit Mode"
          />

          <Button variant="contained" color="primary" disabled={!isEditMode} onClick={() => setIsReviewOpen(true)}>
            Review
          </Button>

          <Button variant="outlined" color="error" onClick={() => navigate("/home")}>
            Close
          </Button>
        </div>
      </div>

      <ReviewDrawer
        open={isReviewOpen}
        changesLog={changesLog}
        isDeploying={isDeploying}
        onClose={() => setIsReviewOpen(false)}
        onDeploy={handleDeployChanges}
        onApplyConfig={(id, config) => {
          updateNodeConfig(id, config);
        }}
      />

      <div
        ref={reactFlowWrapper}
        style={{ height: "92vh", width: "100%", backgroundColor: "#f9f9f9" }}
      >
        <ReactFlow
          nodes={nodeValue}
          edges={edgeValue}
          onNodesChange={updateNodes}
          onEdgesChange={updateEdges}
          onConnect={isEditMode ? onConnect : undefined}
          nodeTypes={nodeTypes}
          onInit={setRFInstance}
          onEdgeClick={onEdgeClick}
          nodesDraggable={isEditMode}
          nodesConnectable={isEditMode}
          elementsSelectable={isEditMode}
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
              <Trash2
                onClick={() => {
                  deleteEdge(selectedEdge);
                  setSelectedEdge(null);
                }}
                className="text-red-500 cursor-pointer"
                size={16}
              />
            </Panel>
          )}
        </ReactFlow>

        <div
          style={{
            position: "absolute",
            bottom: "5rem",
            left: "50%",
            transform: "translateX(-50%)",
            backgroundColor: "#f1f5f9",
            padding: "12px 24px",
            borderRadius: "8px",
            boxShadow: "0 4px 8px rgba(0,0,0,0.1)",
            display: "flex",
            gap: "12px",
            zIndex: 20,
          }}
        >
          <PluginDropdownOptions
            key={`source-${agentType}`}
            kind={agentType === "fluent-bit" ? "input" : "receiver"}
            nodeType="source"
            label="Source"
            dataType={agentType === "fluent-bit" ? "input" : "receiver"}
            disabled={!isEditMode}
            agentType={agentType}
          />

          <PluginDropdownOptions
            key={`processor-${agentType}`}
            kind={agentType === "fluent-bit" ? "filter" : "processor"}
            nodeType="processor"
            label="Processor"
            dataType={agentType === "fluent-bit" ? "filter" : "processor"}
            disabled={!isEditMode}
            agentType={agentType}
          />

          <PluginDropdownOptions
            key={`destination-${agentType}`}
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

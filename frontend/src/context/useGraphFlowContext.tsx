import React, { createContext, useContext } from "react";
import { applyNodeChanges, Node, NodeChange, useNodesState, applyEdgeChanges, Edge, EdgeChange, useEdgesState, addEdge, XYPosition } from "reactflow";
import { initialNodes, initialEdges } from "../constants";

interface BaseNodeData {
	component_role?: string;
	name?: string;
	supported_signals?: string[];
	component_name?: string;
	config?: any;
}

interface NodeData extends BaseNodeData {
	component_id?: string | number;
	position?: XYPosition;
}

interface EdgeData {
	id?: string;
	sourceComponentId?: string | number;
	targetComponentId?: string | number;
}

interface NewNode {
	type: string;
	position: XYPosition;
	data: BaseNodeData;
}

interface GraphFlowContextType {
	nodeValue: Node<NodeData>[];
	edgeValue: Edge<EdgeData>[];
	updateNodes: (changes: NodeChange[]) => void;
	updateEdges: (changes: EdgeChange[]) => void;
	connectNodes: (params: { source: string; target: string }) => void;
	resetGraph: () => void;
	deleteNode: (nodeId: string) => void;
	addNode: (newNode: NewNode) => string;
	updateNodeConfig: (nodeId: string, config: any) => void;
}

// Validation functions
const validateNodeConnection = (sourceNode: Node<NodeData>, targetNode: Node<NodeData>): { isValid: boolean; error?: string } => {
	if (!sourceNode || !targetNode) {
		return { isValid: false, error: "Source or target node not found" };
	}

	if (sourceNode.type === targetNode.type) {
		return { isValid: false, error: "Source and target are of the same type" };
	}

	// Validate flow direction
	const validFlow = 
		(sourceNode.type === "source" && (targetNode.type === "processor" || targetNode.type === "destination")) ||
		(sourceNode.type === "processor" && targetNode.type === "destination");

	if (!validFlow) {
		return { isValid: false, error: "Invalid flow direction" };
	}

	// Check signal compatibility
	const hasCommonSignals = targetNode.data.supported_signals?.some(signal =>
		sourceNode.data.supported_signals?.includes(signal)
	);

	if (!hasCommonSignals) {
		return { isValid: false, error: "Source and target are not compatible" };
	}

	return { isValid: true };
};

const validateAllEdges = (edges: Edge<EdgeData>[], nodes: Node<NodeData>[]): { isValid: boolean; error?: string } => {
	for (const edge of edges) {
		const sourceNode = nodes.find(node => node.id === edge.source);
		const targetNode = nodes.find(node => node.id === edge.target);
		
		const validation = validateNodeConnection(sourceNode!, targetNode!);
		if (!validation.isValid) {
			return validation;
		}
	}
	return { isValid: true };
};

const fetchLocalStorageNodesData = () => {
	try {
		const Nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");

		if (Nodes.length === 0) {
			console.log("LocalStorage is empty. Initializing with initialNodes.");
			localStorage.setItem("Nodes", JSON.stringify(initialNodes));
			return initialNodes;
		}

		const isReactFlowFormat = Nodes.length > 0 && "type" in Nodes[0] && "data" in Nodes[0];
		if (isReactFlowFormat) {
			return Nodes;
		}

		return convertToReactFlowNodesFormat(Nodes);
	} catch (error) {
		console.error("Failed to parse Nodes from localStorage:", error);
		return [];
	}
};

const fetchLocalStorageEdgeData = () => {
	try {
		const Edges = JSON.parse(localStorage.getItem("PipelineEdges") || "[]");

		if (Edges.length === 0) {
			console.log("LocalStorage is empty. Initializing with initialEdges.");
			localStorage.setItem("PipelineEdges", JSON.stringify(initialEdges));
			return initialEdges;
		}

		const isReactFlowFormat = Edges.length > 0 && "source" in Edges[0] && "target" in Edges[0];
		if (isReactFlowFormat) {
			return Edges;
		}

		return convertToReactFlowEdgeFormatEdges(Edges);
	} catch (error) {
		console.error("Failed to parse Edges from localStorage:", error);
		return [];
	}
};

const convertToReactFlowNodesFormat = (data: NodeData[]) => {
	return data.map((source: NodeData, index: number) => ({
		id: source.component_id?.toString() || `${index}`,
		type:
			source.component_role === "receiver"
				? "source"
				: source.component_role === "exporter"
					? "destination"
					: "processor",
		position: source.position || { x: 100, y: 100 + index * 100 },
		data: {
			label: `${source.name || "Unnamed"}-(${index + 1})`,
			component_id: source.component_id?.toString() || `${index}`,
			component_role: source.component_role || "",
			name: source.name || "Unnamed",
			supported_signals: source.supported_signals || [],
			component_name: source.component_name || "",
			config: source.config || {},
		},
	}));
};

const convertToReactFlowEdgeFormatEdges = (data: EdgeData[]) => {
	return data.map((edge: EdgeData, index: number) => ({
		id: edge.id || `edge-${index}`,
		source: edge.sourceComponentId?.toString() || "",
		target: edge.targetComponentId?.toString() || "",
		type: "smoothstep",
		animated: true,
		data: {
			sourceComponentId: edge.sourceComponentId?.toString() || "",
			targetComponentId: edge.targetComponentId?.toString() || "",
		},
	}));
};

const GraphFlowContext = createContext<GraphFlowContextType | undefined>(undefined);

export const GraphFlowProvider = ({ children }: { children: React.ReactNode }) => {
	const [nodeValue, setNodeValue] = useNodesState(fetchLocalStorageNodesData());
	const [edgeValue, setEdgeValue] = useEdgesState(fetchLocalStorageEdgeData());

	// useEffect(() => {
	// 	const handleStorageChange = (event: StorageEvent) => {
	// 		if (event.key === "Nodes" && event.newValue) {
	// 			try {
	// 				const updatedNodes = JSON.parse(event.newValue);
	// 				const isReactFlowFormat = updatedNodes.length > 0 && "type" in updatedNodes[0] && "data" in updatedNodes[0];
	// 				const formattedNodes = isReactFlowFormat ? updatedNodes : convertToReactFlowNodesFormat(updatedNodes);
	// 				setNodeValue(formattedNodes);
	// 			} catch (error) {
	// 				console.error("Error parsing updated Nodes from localStorage:", error);
	// 			}
	// 		}
	// 		if (event.key === "PipelineEdges" && event.newValue) {
	// 			try {
	// 				const updatedEdges = JSON.parse(event.newValue);
	// 				const isReactFlowFormat = updatedEdges.length > 0 && "source" in updatedEdges[0] && "target" in updatedEdges[0];
	// 				const formattedEdges = isReactFlowFormat ? updatedEdges : convertToReactFlowEdgeFormatEdges(updatedEdges);
	// 				setEdgeValue(formattedEdges);
	// 			} catch (error) {
	// 				console.error("Error parsing updated Edges from localStorage:", error);
	// 			}
	// 		}
	// 	};

	// 	window.addEventListener("storage", handleStorageChange);
	// 	return () => {
	// 		window.removeEventListener("storage", handleStorageChange);
	// 	};
	// }, [setNodeValue, setEdgeValue]);

	const updateNodes = (changes: NodeChange[]) => {
		setNodeValue(prevNodes => {
			const updatedNodes = applyNodeChanges(changes, prevNodes);
			localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
			return updatedNodes;
		});
	};

	const updateEdges = (changes: EdgeChange[]) => {
		setEdgeValue(prevEdges => {
			const updatedEdges = applyEdgeChanges(changes, prevEdges);
			const validation = validateAllEdges(updatedEdges, nodeValue);
			
			if (!validation.isValid) {
				console.error("Invalid edge configuration:", validation.error);
				return prevEdges;
			}

			localStorage.setItem("PipelineEdges", JSON.stringify(updatedEdges));
			return updatedEdges;
		});
	};

	const connectNodes = (params: { source: string; target: string }) => {
		const sourceNode = nodeValue.find(node => node.id === params.source);
		const targetNode = nodeValue.find(node => node.id === params.target);

		const validation = validateNodeConnection(sourceNode!, targetNode!);
		if (!validation.isValid) {
			console.error("Invalid connection:", validation.error);
			return;
		}

		setEdgeValue(prevEdges => {
			const edgeId = `edge-${params.source}-${params.target}`;
			const updatedEdges = addEdge(
				{
					id: edgeId,
					source: params.source,
					target: params.target,
					type: "smoothstep",
					animated: true,
					data: {
						sourceComponentId: parseInt(params.source, 10),
						targetComponentId: parseInt(params.target, 10),
					},
				},
				prevEdges
			);
			localStorage.setItem("PipelineEdges", JSON.stringify(updatedEdges));
			return updatedEdges;
		});
	};

	const resetGraph = () => {
		// Reset nodes
		setNodeValue(initialNodes);
		localStorage.setItem("Nodes", JSON.stringify(initialNodes));

		// Reset edges
		setEdgeValue(initialEdges);
		localStorage.setItem("PipelineEdges", JSON.stringify(initialEdges));
	};

	const deleteNode = (nodeId: string) => {
		setNodeValue(prevNodes => {
			const updatedNodes = prevNodes.filter(node => node.id !== nodeId);
			localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
			return updatedNodes;
		});

		setEdgeValue(prevEdges => {
			const updatedEdges = prevEdges.filter(edge => edge.source !== nodeId && edge.target !== nodeId);
			localStorage.setItem("PipelineEdges", JSON.stringify(updatedEdges));
			return updatedEdges;
		});
	};

	const addNode = (newNode: NewNode): string => {
		let newNodeId = '';
		setNodeValue(prevNodes => {
			newNodeId = (prevNodes.length + 1).toString();
			const newNodeToAdd = {
				id: newNodeId,
				type: newNode.type,
				position: newNode.position,
				data: {
					...newNode.data,
					component_id: newNodeId,
				}
			};
			const updatedNodes = [...prevNodes, newNodeToAdd];
			localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
			return updatedNodes;
		});
		return newNodeId;
	};

	const updateNodeConfig = (nodeId: string, config: any) => {
		setNodeValue(prevNodes => {
			const updatedNodes = prevNodes.map(node => node.id === nodeId ? { ...node, data: { ...node.data, config } } : node);
			localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
			return updatedNodes;
		});
	};

	return (
		<GraphFlowContext.Provider value={{ 
			nodeValue, 
			edgeValue, 
			updateNodes, 
			updateEdges, 
			connectNodes,
			resetGraph,
			deleteNode,
			addNode,
			updateNodeConfig
		}}>
			{children}
		</GraphFlowContext.Provider>
	);
};

export const useGraphFlow = () => {
	const context = useContext(GraphFlowContext);
	if (!context) {
		throw new Error("useGraphFlow must be used within a GraphFlowProvider");
	}
	return context;
};

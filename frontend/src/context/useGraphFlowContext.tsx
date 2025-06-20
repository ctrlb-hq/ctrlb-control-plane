import React, { createContext, useContext, useState } from "react";
import {
	applyNodeChanges,
	Node,
	NodeChange,
	useNodesState,
	applyEdgeChanges,
	Edge,
	EdgeChange,
	useEdgesState,
	Connection,
	XYPosition,
} from "reactflow";
import { useToast } from "@/hooks/useToast";
import { capitalize } from "@/utils/utils";

export interface Changes {
	id?: string;
	type: string;
	name: string;
	status: string;
	initialConfig?: any;
	finalConfig?: any;
	component_type?: string;
}

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

export interface GraphFlowContextType {
	nodeValue: Node<NodeData>[];
	edgeValue: Edge<EdgeData>[];
	changesLog: Changes[];
	setNodeValueDirect: (nodes: Node<NodeData>[]) => void;
	setEdgeValueDirect: (edges: Edge<EdgeData>[]) => void;
	updateNodes: (changes: NodeChange[]) => void;
	updateEdges: (changes: EdgeChange[]) => void;
	connectNodes: (params: Edge<EdgeData> | Connection) => void;
	deleteNode: (nodeId: string) => void;
	deleteEdge: (params: Edge<EdgeData> | Connection) => void;
	addNode: (newNode: NewNode) => string;
	updateNodeConfig: (nodeId: string, config: any) => void;
	clearChangesLog: () => void;
	resetGraph: () => void;
}

const GraphFlowContext = createContext<GraphFlowContextType | undefined>(undefined);

export const GraphFlowProvider = ({ children }: { children: React.ReactNode }) => {
	const [nodeValue, setNodeValue] = useNodesState<NodeData>([]);
	const [edgeValue, setEdgeValue] = useEdgesState<EdgeData>([]);
	const [changesLog, setChangesLog] = useState<Changes[]>([]);
	const { toast } = useToast();

	// helper: human-readable node name
	const findNodeName = (id: string) => nodeValue.find(n => n.id === id)?.data.name ?? `#${id}`;

	// upsert change entry by id
	const addChange = (c: Changes & { id?: string }) => {
		setChangesLog(prev => {
			// no id → just append
			if (!c.id) return [...prev, c];

			const idx = prev.findIndex(x => x.id === c.id);
			// found an existing entry
			if (idx >= 0) {
				const existing = prev[idx];
				// if we're re-adding something that was deleted, remove that deletion record
				if (c.status === "added" && existing.status === "deleted") {
					return prev.filter(x => x.id !== c.id);
				}
				// else merge/update as before
				const updated = [...prev];
				updated[idx] = { ...updated[idx], ...c };
				return updated;
			}

			// no prior entry → append new
			return [...prev, c];
		});
	};
	const clearChangesLog = () => setChangesLog([]);

	// NODE OPERATIONS
	const addNode = (newNode: NewNode): string => {
		// generate numeric id not currently used
		const usedIds = nodeValue.map(n => n.id);
		const newId =
			Array.from({ length: usedIds.length + 2 }, (_, i) => (i + 1).toString()).find(
				id => !usedIds.includes(id),
			) || `${usedIds.length + 1}`;

		const nodeToAdd: Node<NodeData> = {
			id: newId,
			type: newNode.type,
			position: newNode.position,
			data: { ...newNode.data, component_id: newId },
		};

		setNodeValue([...nodeValue, nodeToAdd]);

		addChange({
			id: newId,
			type: capitalize(newNode.type),
			name: newNode.data.name || `Node ${newId}`,
			status: "added",
			finalConfig: newNode.data.config,
			component_type: newNode.data.component_name,
		});

		return newId;
	};

	const updateNodeConfig = (nodeId: string, config: any) => {
		const before = nodeValue.find(n => n.id === nodeId);
		setNodeValue(nodeValue.map(n => (n.id === nodeId ? { ...n, data: { ...n.data, config } } : n)));
		if (before) {
			addChange({
				id: nodeId,
				type: capitalize(before.type || "NULL"),
				name: before.data.name || "NULL",
				status: "edited",
				initialConfig: before.data.config,
				finalConfig: config,
				component_type: before.data.component_name,
			});
		}
	};

	const deleteNode = (nodeId: string) => {
		// log and remove edges connected to node
		const edgesToRemove = edgeValue.filter(e => e.source === nodeId || e.target === nodeId);
		edgesToRemove.forEach(edge => {
			addChange({
				id: edge.id!,
				type: "Edge",
				name: `${findNodeName(edge.source)} → ${findNodeName(edge.target)}`,
				status: "deleted",
				initialConfig: { source: edge.source, target: edge.target },
				component_type: "edge",
			});
		});
		setEdgeValue(edgeValue.filter(e => e.source !== nodeId && e.target !== nodeId));

		// log and remove node
		const nodeToDelete = nodeValue.find(n => n.id === nodeId);
		if (nodeToDelete) {
			addChange({
				id: nodeId,
				type: capitalize(nodeToDelete.type || "NULL"),
				name: nodeToDelete.data.name || "",
				status: "deleted",
				initialConfig: nodeToDelete.data.config,
				component_type: nodeToDelete.data.component_name,
			});
		}
		setNodeValue(nodeValue.filter(n => n.id !== nodeId));
	};

	// EDGE OPERATIONS
	const connectNodes = (params: Edge<EdgeData> | Connection) => {
		const { source, target } = params;
		if (!source || !target) {
			toast({ title: "Invalid edge: source or target is missing", variant: "destructive" });
			return;
		}
		const edgeId = `edge-${source}-${target}`;
		const newEdge: Edge<EdgeData> = { id: edgeId, source, target, animated: true };
		if (!edgeValue.some(e => e.id === edgeId)) {
			setEdgeValue([...edgeValue, newEdge]);
			addChange({
				id: edgeId,
				type: "Edge",
				name: `${findNodeName(source)} → ${findNodeName(target)}`,
				status: "added",
				finalConfig: { source, target },
				component_type: "edge",
			});
		}
	};

	const deleteEdge = (params: Edge<EdgeData> | Connection) => {
		const targetEdge = edgeValue.find(e => e.source === params.source && e.target === params.target);
		if (!targetEdge) {
			toast({ title: "Edge not found", variant: "destructive" });
			return;
		}
		setEdgeValue(edgeValue.filter(e => !(e.source === params.source && e.target === params.target)));
		addChange({
			id: targetEdge.id!,
			type: "Edge",
			name: `${findNodeName(targetEdge.source)} → ${findNodeName(targetEdge.target)}`,
			status: "deleted",
			initialConfig: { source: targetEdge.source, target: targetEdge.target },
			component_type: "edge",
		});
	};
	const resetGraph = () => {
		setNodeValue([]);
		setEdgeValue([]);
		clearChangesLog();
	};

	return (
		<GraphFlowContext.Provider
			value={{
				resetGraph,
				nodeValue,
				edgeValue,
				changesLog,
				setNodeValueDirect: setNodeValue,
				setEdgeValueDirect: setEdgeValue,
				updateNodes: changes => setNodeValue(applyNodeChanges(changes, nodeValue)),
				updateEdges: changes => setEdgeValue(applyEdgeChanges(changes, edgeValue)),
				connectNodes,
				deleteNode,
				deleteEdge,
				addNode,
				updateNodeConfig,
				clearChangesLog,
			}}>
			{children}
		</GraphFlowContext.Provider>
	);
};

export const useGraphFlow = () => {
	const ctx = useContext(GraphFlowContext);
	if (!ctx) throw new Error("useGraphFlow must be used within GraphFlowProvider");
	return ctx;
};

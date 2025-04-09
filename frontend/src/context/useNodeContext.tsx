import React, { createContext, Dispatch, SetStateAction, useContext } from "react";
import { Node, NodeChange, useNodesState } from "reactflow";

interface NodeValueContextType {
  nodeValue: Node<any, string | undefined>[];
  setNodeValue: Dispatch<SetStateAction<Node<any, string | undefined>[]>>;
  onNodesChange: (changes: NodeChange[]) => void;
}

// Safely parse Nodes from localStorage
const fetchLocalStorageData = () => {
  try {
    const Nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
    return { Nodes };
  } catch (error) {
    console.error("Failed to parse Nodes from localStorage:", error);
    return { Nodes: [] };
  }
};


const { Nodes } = fetchLocalStorageData()

// Initialize nodes with fallback for missing position
const initialNodes: Node<any, string | undefined>[] = [
  Nodes,
  ...Nodes.map((source: any, index: number) => ({
    id: source.component_id.toString(),
    type: source.component_role == "receiver" ? "source" : source.component_role == "exporter" ? "destination" : "processor",
    position: source.position || { x: 100, y: 100 + index * 100 }, // Ensure position is set
    data: {
      label: (
        <div style={{ fontSize: '10px', textAlign: 'center' }}>
          {`${source.name}-(${index + 1})`}
        </div>
      ),
      id: source.component_id.toString(),
      type: source.component_role,
      name: source.name,
      supported_signals: source.supported_signals,
      component_name: source.component_name,
    },
  })),
];

const NodeValueContext = createContext<NodeValueContextType | undefined>(undefined);

export const NodeValueProvider = ({ children }: { children: React.ReactNode }) => {
  const [nodeValue, setNodeValue, onNodesChange] = useNodesState(initialNodes);
  
  return (
    <NodeValueContext.Provider value={{ nodeValue, setNodeValue, onNodesChange }}>
      {children}
    </NodeValueContext.Provider>
  );
};

export const useNodeValue = () => {
  const context = useContext(NodeValueContext);
  if (!context) {
    throw new Error("useNodeValue must be used within a NodeValueProvider");
  }
  return context;
};
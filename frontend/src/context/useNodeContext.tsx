import React, { createContext, Dispatch, SetStateAction, useContext, useEffect } from "react";
import { applyNodeChanges, Node, NodeChange, useNodesState } from "reactflow";

interface NodeValueContextType {
  nodeValue: Node<any, string | undefined>[];
  setNodeValue: Dispatch<SetStateAction<Node<any, string | undefined>[]>>;
  onNodesChange: (changes: NodeChange[]) => void;
}

const fetchLocalStorageData = () => {
  try {
    const Nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");

    const isReactFlowFormat = Nodes.length > 0 && "type" in Nodes[0] && "data" in Nodes[0];
    if (isReactFlowFormat) {
      return Nodes;
    }

    return convertToNodes(Nodes);
  } catch (error) {
    console.error("Failed to parse Nodes from localStorage:", error);
    return [];
  }
};

const convertToNodes = (data: any[]) => {
  return data.map((source: any, index: number) => ({
    id: source.component_id?.toString() || `${index}`,
    type: source.component_role === "receiver"
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
      config:source.config || {},
    },
  }));
};

const NodeValueContext = createContext<NodeValueContextType | undefined>(undefined);

export const NodeValueProvider = ({ children }: { children: React.ReactNode }) => {
  const [nodeValue, setNodeValue] = useNodesState(fetchLocalStorageData());

  useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === "Nodes" && event.newValue) {
        try {
          const updatedNodes = JSON.parse(event.newValue);

          // Detect if the data is already in ReactFlow format
          const isReactFlowFormat = updatedNodes.length > 0 && "type" in updatedNodes[0] && "data" in updatedNodes[0];
          const formattedNodes = isReactFlowFormat ? updatedNodes : convertToNodes(updatedNodes);

          setNodeValue(formattedNodes);
        } catch (error) {
          console.error("Error parsing updated Nodes from localStorage:", error);
        }
      }
    };

    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [setNodeValue]);

  const onNodesChange = (changes: NodeChange[]) => {
    setNodeValue((prevNodes) => {
      const updatedNodes = applyNodeChanges(changes, prevNodes);

      // Save the updated nodes to localStorage
      localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
      return updatedNodes;
    });
  };

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
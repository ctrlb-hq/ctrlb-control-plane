import React, { createContext, Dispatch, SetStateAction, useContext, useEffect } from "react";
import { applyEdgeChanges, Edge, EdgeChange, useEdgesState } from "reactflow";
import { initialEdges } from "../constants";

interface EdgeValueContextType {
    edgeValue: Edge<any>[];
    setEdgeValue: Dispatch<SetStateAction<Edge<any>[]>>;
    onEdgesChange: (changes: EdgeChange[]) => void;
}

interface EdgeData {
    id?: string;
    sourceComponentId?: string | number;
    targetComponentId?: string | number;
}

const fetchLocalStorageEdgeData = () => {
    try {
        const Edges = JSON.parse(localStorage.getItem("PipelineEdges") || "[]");

        // If localStorage is empty, initialize it with initialEdges
        if (Edges.length === 0) {
            console.log("LocalStorage is empty. Initializing with initialEdges.");
            localStorage.setItem("PipelineEdges", JSON.stringify(initialEdges));
            return initialEdges;
        }

        // Detect if the data is already in ReactFlow format
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

const EdgeValueContext = createContext<EdgeValueContextType | undefined>(undefined);

export const EdgeValueProvider = ({ children }: { children: React.ReactNode }) => {
    const [edgeValue, setEdgeValue] = useEdgesState(fetchLocalStorageEdgeData());

    useEffect(() => {
        const handleStorageChange = (event: StorageEvent) => {
            if (event.key === "PipelineEdges" && event.newValue) {
                try {
                    const updatedEdges = JSON.parse(event.newValue);

                    // Detect if the data is already in ReactFlow format
                    const isReactFlowFormat =
                        updatedEdges.length > 0 && "source" in updatedEdges[0] && "target" in updatedEdges[0];
                    const formattedEdges = isReactFlowFormat ? updatedEdges : convertToReactFlowEdgeFormatEdges(updatedEdges);

                    setEdgeValue(formattedEdges);
                } catch (error) {
                    console.error("Error parsing updated Edges from localStorage:", error);
                }
            }
        };

        window.addEventListener("storage", handleStorageChange);

        return () => {
            window.removeEventListener("storage", handleStorageChange);
        };
    }, []);

    const onEdgesChange = (changes: EdgeChange[]) => {
        setEdgeValue(prevEdges => {
            const updatedEdges = applyEdgeChanges(changes, prevEdges);

            // Save the updated edges to localStorage
            localStorage.setItem("PipelineEdges", JSON.stringify(updatedEdges));
            return updatedEdges;
        });
    };

    return (
        <EdgeValueContext.Provider value={{ edgeValue, setEdgeValue, onEdgesChange }}>
            {children}
        </EdgeValueContext.Provider>
    );
};

export const useEdgeValue = () => {
    const context = useContext(EdgeValueContext);
    if (!context) {
        throw new Error("useEdgeValue must be used within an EdgeValueProvider");
    }
    return context;
};

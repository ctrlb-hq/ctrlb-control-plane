import { Edge, Node } from "reactflow";
export interface PipelineNodeData {
    component_id: string;
    name: string;
    component_name: string;
    component_type: string;
    supported_signals: string[];
    config: unknown;
}

export const initialNodes: Node<PipelineNodeData>[] = [

    {
        "id": "1",
        "type": "destination",
        "position": {
            "x": -1,
            "y": -1
        },
        "data": {
            "component_id": "1",
            "name": "Debug Exporter Configuration",
            "component_name": "debug_exporter",
            "component_type": "exporter",
            "supported_signals": ["traces", "metrics", "logs"],
            "config": {
                "format": "json"
            }
        },
        "width": 120,
        "height": 64,
        "selected": false,
        "dragging": false,
        "positionAbsolute": {
            "x": -1,
            "y": -1
        }
    },

    {
        "id": "2",
        "type": "source",
        "position": {
            "x": -1,
            "y": -1
        },
        "data": {
            "component_type": "receiver",
            "component_id": "2",
            "name": "OTLP Receiver Configuration",
            "supported_signals": [
                "traces",
                "metrics",
                "logs"
            ],
            "component_name": "otlp_receiver",
            "config": {
                "protocols": {
                    "http": {
                        "endpoint": "0.0.0.0:4317"
                    }
                }
            }
        },
        "width": 134,
        "height": 64,
        "selected": false,
        "dragging": false,
        "positionAbsolute": {
            "x": -1,
            "y": -1
        }
    }
];

export const initialEdges: Edge[] = [
    {
        "source": "2",
        "sourceHandle": null,
        "target": "1",
        "targetHandle": null,
        "animated": true,
        "data": {
            "sourceComponentId": 2,
            "targetComponentId": 1
        },
        "id": "reactflow__edge-2-1"
    }
];
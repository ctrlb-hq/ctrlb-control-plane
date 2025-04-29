import { Edge, Node } from "reactflow";

export const initialNodes: Node[] = [

    {
        "id": "1",
        "type": "source",
        "position": {
            "x": 166.56891869734653,
            "y": 368.60026536267003
        },
        "data": {
            "component_id": "1",
            "name": "Debug Exporter Configuration",
            "component_name": "debug_exporter",
            "supported_signals": [],
            "config": {
                "format": "json"
            }
        },
        "width": 120,
        "height": 64,
        "selected": false,
        "dragging": false,
        "positionAbsolute": {
            "x": 166.56891869734653,
            "y": 368.60026536267003
        }
    },

    {
        "id": "2",
        "type": "destination",
        "position": {
            "x": -120.99553206402538,
            "y": 346.4567463022786
        },
        "data": {
            "label": {
                "type": "div",
                "key": null,
                "ref": null,
                "props": {
                    "style": {
                        "fontSize": "10px",
                        "textAlign": "center"
                    },
                    "children": "OTLP Receiver Configuration-(3)"
                },
                "_owner": null,
                "_store": {}
            },
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
        "selected": true,
        "positionAbsolute": {
            "x": -120.99553206402538,
            "y": 346.4567463022786
        },
        "dragging": false
    }
];

export const initialEdges: Edge[] = [
    // { id: 'e1-2', source: 'system', target: 'mask_ssn', label: '11GB', animated: true },
    // { id: 'e2-2', source: 'mask_ssn', target: 'ctrlB', label: '1MB', animated: true },
    {
        "source": "1",
        "sourceHandle": null,
        "target": "2",
        "targetHandle": null,
        "animated": true,
        "data": {
            "sourceComponentId": 3,
            "targetComponentId": 2
        },
        "id": "reactflow__edge-1-2"
    }
];
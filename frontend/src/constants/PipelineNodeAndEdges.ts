import { Edge, Node } from "reactflow";

export const initialNodes: Node[] = [
    {
        id: 'system',
        type: 'source',
        position: { x: 50, y: 200 },

        data: {
            label: 'system',
            supported_signals: [
                'LOG'
            ],
            type:'source',
            name:'system',
            sublabel: 'fluentbit',

        },
    },

    {
        id: 'mask_ssn',
        type: 'processor',
        position: { x: 250, y: 200 },
        data: {
            label: 'mask_ssn',
            supported_signals:[
                'LOG'
            ],
            type:"processor",
            name:"mask_ssn",
            sublabel: 'mask',
        },
    },

    // Destinations
    {
        id: 'ctrlB',
        type: 'destination',
        position: { x: 750, y: 80 },
        data: {
            label: 'ctrlB',
            type:"destination",
            name:"ctrlb",
            supported_signals:[
                'LOG'
            ]
        },
    },
];

export const initialEdges: Edge[] = [
    { id: 'e1-2', source: 'system', target: 'mask_ssn', label: '11GB', animated: true },
    { id: 'e2-2', source: 'mask_ssn', target: 'ctrlB', label: '1MB', animated: true },
];
import { Edge, Node } from "reactflow";

export const initialNodes: Node[] = [
    // Source
    {
        id: 'system',
        type: 'source',
        position: { x: 50, y: 200 },
        data: {
            label: 'system',
            sublabel: 'fluentbit',
            inputType: '',
            outputType: 'LOG',
            icon: '→|'
        },
    },
    // Processors
    {
        id: 'mask_ssn',
        type: 'processor',
        position: { x: 250, y: 200 },
        data: {
            label: 'mask_ssn',
            sublabel: 'mask',
            inputType: 'LOG',
            outputType: 'LOG'
        },
    },
    {
        id: 'drop_trace',
        type: 'processor',
        position: { x: 500, y: 80 },
        data: {
            label: 'drop_trace',
            sublabel: 'regex_filter',
            inputType: 'LOG',
            outputType: 'LOG'
        },
    },
    {
        id: 'error_monitor',
        type: 'processor',
        position: { x: 500, y: 200 },
        data: {
            label: 'error_monitor',
            sublabel: 'log_to_metric',
            inputType: 'LOG',
            outputType: 'METRIC'
        },
    },
    {
        id: 'exception_m',
        type: 'processor',
        position: { x: 500, y: 320 },
        data: {
            label: 'exception_m',
            sublabel: 'log_to_metric',
            inputType: 'LOG',
            outputType: 'METRIC'
        },
    },
    {
        id: 'log_to_pattern',
        type: 'processor',
        position: { x: 500, y: 440 },
        data: {
            label: 'log_to_pattern',
            sublabel: 'log_to_pattern',
            inputType: 'LOG',
            outputType: 'PATTERN & SAMPLE'
        },
    },
    // Destinations
    {
        id: 'ctrlB',
        type: 'destination',
        position: { x: 750, y: 80 },
        data: {
            label: 'ctrlB',
            sublabel: 'CtrlB_Explore',
            inputType: 'MIXED',
            outputType: ''
        },
    },
    {
        id: 'openmetrics',
        type: 'destination',
        position: { x: 750, y: 320 },
        data: {
            label: 'openmetrics',
            sublabel: 'openmetrics',
            inputType: 'MIXED',
            outputType: ''
        },
    },
];

export const initialEdges: Edge[] = [
    { id: 'e1-2', source: 'system', target: 'mask_ssn', label: '11GB', animated: true },
    { id: 'e2-3', source: 'mask_ssn', target: 'drop_trace', label: '2KB', animated: true },
    { id: 'e2-4', source: 'mask_ssn', target: 'error_monitor', label: '2KB', animated: true },
    { id: 'e2-5', source: 'mask_ssn', target: 'exception_m', label: '2KB', animated: true },
    { id: 'e2-6', source: 'mask_ssn', target: 'log_to_pattern', label: '2KB', animated: true },
    { id: 'e3-7', source: 'drop_trace', target: 'ctrlB', label: '2KB', animated: true },
    { id: 'e4-7', source: 'error_monitor', target: 'openmetrics', label: '1MB', animated: true },
    { id: 'e5-7', source: 'exception_m', target: 'openmetrics', label: '685KB', animated: true },
    { id: 'e6-7', source: 'log_to_pattern', target: 'openmetrics', label: '1MB', animated: true },
];
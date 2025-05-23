import { Edge, Node } from "reactflow";

export interface Changes {
	type: string;
	name: string;
	status: string;
	
}

export interface PipelineNodeData {
	component_id: string;
	name: string;
	component_name: string;
	component_type: string;
	supported_signals: string[];
	config: unknown;
}

export const ROUTES = {
	LOGIN: "/login",
	REGISTER: "/register",
	HOME: "/home",
	CONFIG: "/config",
} as const;

export const steps = [
	{
		title: "Install Pipeline",
		description: "Specify basic settings for pipeline",
	},
	{
		title: "Configure Pipeline",
		description: "Add sources, processor, destination to your pipeline.",
	},
];

export const initialNodes: Node<PipelineNodeData>[] = [
	{
		id: "1",
		type: "destination",
		position: {
			x: 400,
			y: 100,
		},
		data: {
			component_id: "1",
			name: "Debug Exporter Configuration",
			component_name: "debug_exporter",
			component_type: "exporter",
			supported_signals: ["traces", "metrics", "logs"],
			config: {
				format: "json",
			},
		},
		width: 120,
		height: 64,
		selected: false,
		dragging: false,
		positionAbsolute: {
			x: 400,
			y: 100,
		},
	},

	{
		id: "2",
		type: "source",
		position: {
			x: 100,
			y: 100,
		},
		data: {
			component_type: "receiver",
			component_id: "2",
			name: "OTLP Receiver Configuration",
			supported_signals: ["traces", "metrics", "logs"],
			component_name: "otlp_receiver",
			config: {
				protocols: {
					http: {
						endpoint: "0.0.0.0:4317",
					},
				},
			},
		},
		width: 134,
		height: 64,
		selected: false,
		dragging: false,
		positionAbsolute: {
			x: 100,
			y: 100,
		},
	},
];

export const initialEdges: Edge[] = [
	{
		source: "2",
		sourceHandle: null,
		target: "1",
		targetHandle: null,
		animated: true,
		data: {
			sourceComponentId: 2,
			targetComponentId: 1,
		},
		id: "edge-2-1",
	},
];

export const getRandomChartColor = (name: string) => {
	const colors = ["brown", "gold", "green", "red", "purple", "orange", "blue", "pink", "gray"];
	const charSum = name.split("").reduce((sum, char) => sum + char.charCodeAt(0), 0);
	return colors[charSum % colors.length];
};

export const formatTimestampWithDate = (timestamp: number | undefined) => {
	if (!timestamp) return "N/A";
	const date = new Date(timestamp * 1000); // Convert seconds to milliseconds
	const day = date.getDate().toString().padStart(2, "0");
	const month = (date.getMonth() + 1).toString().padStart(2, "0");
	const year = date.getFullYear();
	const hours = date.getHours().toString().padStart(2, "0");
	const minutes = date.getMinutes().toString().padStart(2, "0");
	const seconds = date.getSeconds().toString().padStart(2, "0");
	return `${day}/${month}/${year} ${hours}:${minutes}:${seconds}`;
};


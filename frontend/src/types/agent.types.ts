export interface ApiError {
	message: string;
	error?: string;
}

export interface agentVal {
	id: string;
	name: string;
	version: string;
	pipeline_id: string;
	pipeline_name: string;
	status: string;
	hostname: string;
	platform: string;
	ip: string;
	labels: { [key: string]: string };
}

export interface Agent {
	id: number;
	name: string;
	status: "connected" | "disconnected" | "pending" | string;
	pipeline_name: string;
	version: string;
	type: string;
	log_rate: number;
	metrics_rate: number;
	trace_rate: number;
	_: string;
	selected?: boolean;
}

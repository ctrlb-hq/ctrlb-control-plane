export interface Agents {
	id: string;
	name: string;
	status: string;
	pipeline_name: string;
	version: string;
	log_rate: number;
	metrics_rate: number;
	trace_rate: number;
	selected?: boolean;
}

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

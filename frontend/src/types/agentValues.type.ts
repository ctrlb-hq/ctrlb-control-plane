export interface AgentValuesTable {
	id: string;
	name: string;
	version: string;
	status: string;
	selected: boolean;
	pipeline_name: string;
	log_rate: number;
	metrics_rate: number;
	trace_rate: number;
}

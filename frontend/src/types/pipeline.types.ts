// export interface Pipeline {
//     id: string;
//     name: string;
//     type: string;
//     version: string;
//     hostname: string;
//     platform: string;
//     config: Config;
//     isPipeline: boolean;
//     registeredAt: string;
//   }

export interface PipelineList {
	id: string;
	name: string;
	agents: number;
	incoming_bytes: string;
	outgoing_bytes: string;
	updatedAt: string;
}

export interface Pipeline {
	id: string;
	name: string;
	created_by: string;
	created_at: number;
	updated_at: number;
}

export interface PipelineOverviewInterface {
	id: string;
	name: string;
	status: string;
	agent_id: number;
	agent_version: string;
	ip_address: string;
	hostname: string;
	platform: string;
	created_at: number;
	created_by: string;
	labels: Record<string, string>;
	config: {
		exporters: Record<string, ExporterConfig>;
		processors: Record<string, ProcessorConfig>;
		receivers: Record<string, ReceiverConfig>;
		service: {
			pipelines: Record<string, {
				exporters: string[];
				processors: string[] | null;
				receivers: string[];
			}>;
		};
		telemetry: {
			metrics: {
				level: string;
				readers: string[];
			};
		};
	};
}

export interface ExporterConfig {
	verbosity?: string;
	[key: string]: unknown;
}

export interface ProcessorConfig {
	[key: string]: unknown;
}

export interface ReceiverConfig {
	protocols?: Record<string, unknown>;
	[key: string]: unknown;
}


// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface Config {}

export interface ApiError {
	message: string;
}

export interface DataPoint {
	timestamp: number;
	value: number;
}

export interface MetricData {
	metric_name: string;
	data_points: DataPoint[];
}

export interface FormSchema {
	title?: string;
	type?: string;
	properties?: Record<string, any>;
	required?: string[];
	[key: string]: any;
}

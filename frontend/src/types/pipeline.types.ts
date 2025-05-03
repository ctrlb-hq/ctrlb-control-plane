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

export interface PipeLineOverview {
	label: string;
	value: string | any[];
}

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface Config {}

export interface ApiError {
	message: string;
}

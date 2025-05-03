export interface PipelineNodeType {
	id: string;
	type: string;
	details: string;
	position?: {
		x: number;
		y: number;
	};
	data: {
		label: string;
		sublabel: string;
		inputType: string;
		outputType: string;
		icon: string;
	};
}

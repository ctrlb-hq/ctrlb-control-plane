import ViewPipelineDetails from "./ViewPipelineDetails";

const ExistingPipelineOverview = ({ pipelineId }: { pipelineId: string }) => {
	return (
		<div>
			<ViewPipelineDetails pipelineId={pipelineId} />
		</div>
	);
};

export default ExistingPipelineOverview;

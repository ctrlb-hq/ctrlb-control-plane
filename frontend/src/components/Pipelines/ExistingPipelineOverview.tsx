import ViewPipelineDetails from "./ViewPipelineDetails";
import { NodeValueProvider } from "@/context/useNodeContext";

const ExistingPipelineOverview = ({ pipelineId }: { pipelineId: string }) => {
	return (
		<div>
			<NodeValueProvider>
				<ViewPipelineDetails pipelineId={pipelineId} />
			</NodeValueProvider>
		</div>
	);
};

export default ExistingPipelineOverview;

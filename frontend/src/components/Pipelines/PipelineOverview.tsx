import PipelineDetails from "./PipelineDetails";
import { NodeValueProvider } from "@/context/useNodeContext";

const PipelineOverview = ({ pipelineId }: { pipelineId: string }) => {
	return (
		<div>
			<NodeValueProvider>
				<PipelineDetails pipelineId={pipelineId} />
			</NodeValueProvider>
		</div>
	);
};

export default PipelineOverview;

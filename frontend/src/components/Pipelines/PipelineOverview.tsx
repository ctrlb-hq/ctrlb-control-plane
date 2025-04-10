import PipelineOverviewTable from './PipelineOverviewTable'
import PipelineDetails from './PipelineDetails'
import { NodeValueProvider } from '@/context/useNodeContext'


const PipelineOverview = ({pipelineId}:{pipelineId:string}) => {
    return (
        <div>
            <NodeValueProvider>
            <PipelineDetails pipelineId={pipelineId}/>

            </NodeValueProvider>
            <PipelineOverviewTable pipelineId={pipelineId} />
        </div>
    )
}

export default PipelineOverview

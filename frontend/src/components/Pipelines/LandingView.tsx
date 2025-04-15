import { usePipelineStatus } from '@/context/usePipelineStatus';

const LandingView = () => {
    const pipelineStatus = usePipelineStatus();
    if (!pipelineStatus) {
        return null;
    }

    return (
        <div className="flex flex-col gap-2 justify-center items-center">
            <p className='font-bold text-xl mt-[6rem]'>Get started</p>
            <p className='text-gray-700'>Create Your First Pipeline</p>
            <p className='text-gray-700'>Pipelines collect data from the sources in the pipeline and route them to desired destination.</p>
        </div>
    );
};

export default LandingView;
import { steps } from '@/constants/ProgressFlowPipelineStep';
import { usePipelineStatus } from '@/context/usePipelineStatus';
import { motion } from 'framer-motion';
const ProgressFlow = () => {

    const pipelineStatus = usePipelineStatus();

    if (!pipelineStatus) {
        return null;
    }

    const { currentStep } = pipelineStatus;

    return (
        <div className="flex flex-1 h-[45rem] p-6 bg-gray-50 shadow-lg rounded-lg">
            <div className="relative">
                {steps.map((step, index) => (
                    <div key={index} className="flex items-start space-x-4">
                        <div className="relative flex flex-col items-center">
                            <div
                                className={`w-4 h-4 rounded-full border-2 ${index <= currentStep ? "border-blue-500 bg-blue-500" : "border-gray-300"
                                    }`}
                            />
                            {index < steps.length - 1 && (
                                <div className="w-px h-12 bg-gray-300 absolute top-4 left-1/2 transform -translate-x-1/2"></div>
                            )}
                        </div>
                        {/* Step Content */}
                        <motion.div
                            className="pb-6"
                            initial={{ opacity: 0, x: -10 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ duration: 0.3 }}
                        >
                            <h3 className={`text-lg font-semibold ${index === currentStep ? "text-blue-600" : "text-gray-700"}`}>
                                {step.title}
                            </h3>
                            <p className="text-gray-500 text-sm">{step.description}</p>
                        </motion.div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export default ProgressFlow

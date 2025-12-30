import { steps } from "@/constants";
import { motion } from "framer-motion";

interface ProgressFlowProps {
  currentStep: number;
}

const ProgressFlow: React.FC<ProgressFlowProps> = ({ currentStep }) => {
  return (
    <div className="flex flex-1 p-6 bg-white border rounded-xl shadow-sm ml-2">
      <div className="relative space-y-8">
        {steps.map((step, index) => {
          const isCompleted = index < currentStep;
          const isActive = index === currentStep;

          return (
            <div key={index} className="flex items-start gap-4 relative">
              <div className="relative flex flex-col items-center">
                <motion.div
                  initial={false}
                  animate={{
                    backgroundColor: isCompleted || isActive ? "#2563eb" : "#fff",
                    borderColor: isCompleted || isActive ? "#2563eb" : "#d1d5db",
                    scale: isActive ? 1.15 : 1,
                  }}
                  transition={{ type: "spring", stiffness: 300, damping: 20 }}
                  className="w-4 h-4 rounded-full border-2 z-10"
                />
                {index < steps.length - 1 && (
                  <div className="relative h-12 w-px mt-1 bg-gray-200 overflow-hidden">
                    <motion.div
                      initial={{ height: 0 }}
                      animate={{ height: isCompleted ? "100%" : "0%" }}
                      transition={{ duration: 0.4 }}
                      className="absolute top-0 left-0 w-full bg-blue-500"
                    />
                  </div>
                )}
              </div>
              <motion.div
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.25 }}
                className={`pb-2 ${
                  isActive ? "text-blue-600" : "text-gray-700"
                }`}
              >
                <h3
                  className={`text-sm font-semibold tracking-wide ${
                    isActive ? "text-blue-600" : "text-gray-800"
                  }`}
                >
                  {step.title}
                </h3>
                <p className="text-xs text-gray-500 leading-relaxed max-w-xs">
                  {step.description}
                </p>
              </motion.div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProgressFlow;

import React, { createContext, useContext, useState } from "react";

interface PipelineProps {
	currentStep: number;
	setCurrentStep: (step: number) => void;
}
const PipelineStatusContext = createContext<PipelineProps | null>(null);

export const PipelineStatusProvider = ({ children }: { children: React.ReactNode }) => {
	const [currentStep, setCurrentStep] = useState<number>(0);
	return (
		<PipelineStatusContext.Provider value={{ currentStep, setCurrentStep }}>
			{children}
		</PipelineStatusContext.Provider>
	);
};

export const usePipelineStatus = () => {
	const context = useContext(PipelineStatusContext);
	return context;
};

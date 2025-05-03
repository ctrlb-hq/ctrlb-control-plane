import { PipelineList } from "@/types/pipeline.types";
import { createContext, useContext, useState, ReactNode } from "react";

interface PipelineOverviewContextProps {
	pipelineOverview: PipelineList | null;
	setPipelineOverview: (overview: PipelineList) => void;
}

const PipelineOverviewContext = createContext<PipelineOverviewContextProps | undefined>(undefined);

export const PipelineOverviewProvider = ({ children }: { children: ReactNode }) => {
	const [pipelineOverview, setPipelineOverview] = useState<PipelineList | null>(null);

	return (
		<PipelineOverviewContext.Provider value={{ pipelineOverview, setPipelineOverview }}>
			{children}
		</PipelineOverviewContext.Provider>
	);
};

export const usePipelineOverview = () => {
	const context = useContext(PipelineOverviewContext);
	if (context === undefined) {
		throw new Error("usePipelineOverview must be used within a PipelineOverviewProvider");
	}
	return context;
};

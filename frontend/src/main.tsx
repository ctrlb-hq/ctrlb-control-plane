import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { PipelineStatusProvider } from "./context/usePipelineStatus.tsx";
import { Toaster } from "./components/ui/toaster.tsx";
import { PipelineOverviewProvider } from "./context/usePipelineDetailContext.tsx";
import { PipelineChangesLogProvider } from "./context/usePipelineChangesLog.tsx";
import { EdgeValueProvider } from "./context/useEdgeContext.tsx";
import { NodeValueProvider } from "./context/useNodeContext.tsx";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<PipelineChangesLogProvider>
			{/* <NodeValueProvider> */}
			<PipelineStatusProvider>
				<PipelineOverviewProvider>
					<NodeValueProvider>
						<EdgeValueProvider>
							<App />
						</EdgeValueProvider>
					</NodeValueProvider>
				</PipelineOverviewProvider>
			</PipelineStatusProvider>
			<Toaster />
		</PipelineChangesLogProvider>
	</StrictMode>,
);

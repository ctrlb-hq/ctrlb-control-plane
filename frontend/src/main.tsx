import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { PipelineStatusProvider } from "./context/usePipelineStatus.tsx";
import { Toaster } from "./components/ui/toaster.tsx";
import { PipelineOverviewProvider } from "./context/usePipelineDetailContext.tsx";
import { PipelineChangesLogProvider } from "./context/usePipelineChangesLog.tsx";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<PipelineChangesLogProvider>
			{/* <NodeValueProvider> */}
			<PipelineStatusProvider>
				<PipelineOverviewProvider>
					<App />
				</PipelineOverviewProvider>
			</PipelineStatusProvider>
			<Toaster />
		</PipelineChangesLogProvider>
	</StrictMode>,
);

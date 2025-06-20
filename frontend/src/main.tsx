import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { Toaster } from "./components/ui/toaster.tsx";
import { PipelineOverviewProvider } from "./context/usePipelineDetailContext.tsx";
import { GraphFlowProvider } from "./context/useGraphFlowContext.tsx";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<PipelineOverviewProvider>
			<GraphFlowProvider>
				<App />
			</GraphFlowProvider>
		</PipelineOverviewProvider>
		<Toaster />
	</StrictMode>,
);

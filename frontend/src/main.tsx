import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { PipelineOverviewProvider } from "./context/usePipelineDetailContext.tsx";
import { GraphFlowProvider } from "./context/useGraphFlowContext.tsx";
import { GlobalSnackbarProvider } from "./context/useGlobalSnackbar.tsx";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<GlobalSnackbarProvider>
			<PipelineOverviewProvider>
				<GraphFlowProvider>
					<App />
				</GraphFlowProvider>
			</PipelineOverviewProvider>
		</GlobalSnackbarProvider>
	</StrictMode>
);

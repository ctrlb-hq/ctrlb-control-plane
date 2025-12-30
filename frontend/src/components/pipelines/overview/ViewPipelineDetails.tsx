import { Boxes } from "lucide-react";
import { useEffect, useState, useCallback } from "react";
import pipelineServices from "@/services/pipeline";
import { PipelineOverviewInterface } from "@/types/pipeline.types";
import "reactflow/dist/style.css";
import PipelinYAML from "./YamlViewer";
import PipelineOverview from "./PipelineOverview";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import DeletePipelineDialog from "./DeletePipelineDialog";
import { Button, Tabs, Tab, Box } from "@mui/material";
import { useNavigate } from "react-router-dom";

const ViewPipelineDetails = ({ pipelineId }: { pipelineId: string }) => {
	const [isOpen, setIsOpen] = useState(false);
	const [pipelineOverviewData, setPipelineOverviewData] = useState<PipelineOverviewInterface>();
	const [tabs, setTabs] = useState<string>("overview");
	const { showSnackbar } = useGlobalSnackbar();
	const navigate = useNavigate();;

	const handleGetPipelineOverview = useCallback(async () => {
		try {
			const response = await pipelineServices.getPipelineOverviewById(pipelineId);
			setPipelineOverviewData(response);
		} catch (error) {
			console.error("Error fetching pipeline overview:", error);
			showSnackbar("Failed to fetch pipeline overview", "error");
		}
	}, [pipelineId, showSnackbar]);


	useEffect(() => {
		handleGetPipelineOverview();
	}, [handleGetPipelineOverview]);

	return (
		<div className="flex flex-col h-[100vh] overflow-hidden">
			{/* Header */}
			<div className="flex items-center justify-between px-6 border-b pb-2 bg-white flex-shrink-0">
				<div className="flex gap-2 items-center">
					<Boxes className="text-gray-700" size={32} />
					<h1 className="text-xl text-gray-800 font-semibold">{pipelineOverviewData?.name}</h1>
				</div>
				<div className="flex items-center w-full md:w-auto">
					<div className="flex gap-2 justify-between w-full">
						<div className="flex gap-2">
							<Button
								variant="contained"
								color="primary"
								onClick={() => {
									navigate(`/pipelines/${pipelineId}/edit`, {
										state: { pipelineName: pipelineOverviewData?.name },
									});
								}}
								sx={{ textTransform: "none" }}
							>
								View / Edit Pipeline
							</Button>
							<DeletePipelineDialog
								isOpen={isOpen}
								setIsOpen={setIsOpen}
								pipelineOverview={pipelineOverviewData}
							/>
						</div>
					</div>
				</div>
			</div>
			<Box sx={{ borderBottom: 1, borderColor: "divider" }}>
				<Tabs
					value={tabs}
					onChange={(_, newValue) => setTabs(newValue)}
					textColor="primary"
					indicatorColor="primary"
				>
					<Tab label="Overview" value="overview" />
					<Tab label="YAML" value="yaml" />
				</Tabs>
			</Box>
			{/* Main Content */}
			<div className="flex-1 overflow-auto mt-4">
				{tabs == "overview" && (
					<>
						<PipelineOverview pipelineId={pipelineId} />
					</>
				)}
				{tabs == "yaml" && <PipelinYAML jsonforms={pipelineOverviewData?.config} />}
			</div>
		</div>
	);
};

export default ViewPipelineDetails;

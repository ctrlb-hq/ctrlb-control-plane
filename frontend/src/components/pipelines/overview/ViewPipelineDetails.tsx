import { Boxes } from "lucide-react";
import { useEffect, useState } from "react";
import { useToast } from "@/hooks/useToast";
import pipelineServices from "@/services/pipeline";

import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";

import "reactflow/dist/style.css";
import PipelinYAML from "./YamlViewer";
import PipelineOverview from "./PipelineOverview";

import DeletePipelineDialog from "./DeletePipelineDialog";
import PipelineEditorSheet from "../editor/PipelineGraphEditor";
import { Button } from "@/components/ui/button";

const ViewPipelineDetails = ({ pipelineId }: { pipelineId: string }) => {
	const [isOpen, setIsOpen] = useState(false);
	const [pipelineOverviewData, setPipelineOverviewData] = useState<any>(null);
	const { toast } = useToast();
	const [tabs, setTabs] = useState<string>("overview");

	const handleGetPipelineOverview = async () => {
		try {
			const response = await pipelineServices.getPipelineOverviewById(pipelineId);
			setPipelineOverviewData(response);
		} catch (error) {
			console.error("Error fetching pipeline overview:", error);
			toast({
				title: "Error",
				description: "Failed to fetch pipeline overview",
				variant: "destructive",
			});
		}
	};

	useEffect(() => {
		handleGetPipelineOverview();
	}, [pipelineId]);

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
							<Sheet>
								<SheetTrigger asChild>
									<Button className="bg-blue-500">View/Edit Pipeline</Button>
								</SheetTrigger>
								<SheetContent className="w-full sm:max-w-full p-0" side="right">
									<PipelineEditorSheet
										pipelineId={pipelineId}
										name={pipelineOverviewData?.name}
										setIsSheetOpen={setIsOpen}
									/>
								</SheetContent>
							</Sheet>
							<DeletePipelineDialog
								isOpen={isOpen}
								setIsOpen={setIsOpen}
								pipelineOverview={pipelineOverviewData}
							/>
						</div>
					</div>
				</div>
			</div>
			<div>
				<ul className="flex border-b">
					<li
						className={`mr-6 cursor-pointer py-2 ${tabs === "overview" ? "border-b-2 border-blue-500" : ""}`}
						onClick={() => setTabs("overview")}>
						Overview
					</li>
					<li
						className={`mr-6 cursor-pointer py-2 ${tabs === "yaml" ? "border-b-2 border-blue-500" : ""}`}
						onClick={() => setTabs("yaml")}>
						YAML
					</li>
				</ul>
			</div>
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

import { Button } from "../../ui/button";
import { PlusIcon } from "lucide-react";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogFooter,
	DialogTitle,
	DialogDescription,
} from "@/components/ui/dialog";
import AddPipelineDetails from "@/components/pipelines/create/AddPipelineDetails";
import { useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import PipelineEditorSheet from "@/components/pipelines/editor/PipelineGraphEditor";

const AddPipelineSheet = () => {
	const [currentStep, setCurrentStep] = useState<number>(0);
	const [isSheetOpen, setIsSheetOpen] = useState<boolean>(false);
	const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);
	const [pipelineId, setPipelineId] = useState<string>("");
	const [pipelineName, setPipelineName] = useState<string>("");

	const { resetGraph, changesLog } = useGraphFlow();

	const handleDialogOkay = () => {
		if (currentStep === 0) {
			// step 0: new pipeline → clear and close
			localStorage.removeItem("Sources");
			localStorage.removeItem("Destination");
			localStorage.removeItem("pipelinename");
			localStorage.removeItem("selectedAgentIds");
			localStorage.removeItem("changesLog");
			localStorage.removeItem("platform");
			resetGraph();
			setCurrentStep(0);
			setIsSheetOpen(false);
		} else {
			resetGraph();
			setIsSheetOpen(false);
		}
		setIsDialogOpen(false);
	};

	const handleDialogCancel = () => {
		setIsDialogOpen(false);
	};

	const getDataFromChild = (id: string, name: string) => {
		setPipelineId(id);
		setPipelineName(name);
	};

	const shouldShowDialog = () => {
		return currentStep === 0 || changesLog.length > 0;
	};

	return (
		<div className="flex flex-col gap-7 justify-center items-center">
			<Sheet
				open={isSheetOpen}
				onOpenChange={open => {
					if (!open) {
						if (shouldShowDialog()) {
							setIsDialogOpen(true);
						} else {
							resetGraph();
							setCurrentStep(0);
							setIsSheetOpen(false);
						}
					} else {
						setIsSheetOpen(true);
					}
				}}>
				<SheetTrigger asChild>
					<Button className="flex gap-1 px-4 py-1 bg-blue-500 text-white" variant="outline">
						<PlusIcon className="h-4 w-4" />
						Add New Pipeline
					</Button>
				</SheetTrigger>
				<SheetContent className={currentStep === 0 ? "" : "w-screen"}>
					{currentStep === 0 ? (
						<AddPipelineDetails
							sendPipelineDataToParent={getDataFromChild}
							currentStep={currentStep}
							setCurrentStep={setCurrentStep}
						/>
					) : (
						<PipelineEditorSheet
							pipelineId={pipelineId}
							name={pipelineName}
							setIsSheetOpen={setIsSheetOpen}
							isEditModeStart={true}
						/>
					)}
				</SheetContent>
			</Sheet>

			{shouldShowDialog() && (
				<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
					<DialogContent className="w-[50rem]">
						<DialogHeader>
							<DialogTitle>
								{currentStep === 0 ? "Discard New Pipeline?" : "Discard Pipeline Edits?"}
							</DialogTitle>
							<DialogDescription>
								{currentStep === 0
									? "All your new pipeline details will be lost. Continue?"
									: "Your graph changes will be lost and you’ll go back to the pipeline details step."}
							</DialogDescription>
						</DialogHeader>
						<DialogFooter>
							<Button variant="outline" onClick={handleDialogCancel}>
								Cancel
							</Button>
							<Button className="bg-blue-500" onClick={handleDialogOkay}>
								OK
							</Button>
						</DialogFooter>
					</DialogContent>
				</Dialog>
			)}
		</div>
	);
};

export default AddPipelineSheet;

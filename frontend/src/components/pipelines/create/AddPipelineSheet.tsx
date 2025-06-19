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

	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [isDialogOpen, setIsDialogOpen] = useState(false);
	const [pipelineId, setPipelineId] = useState<string>("");
	const [pipelineName, setPipelineName] = useState<string>("");
	const { resetGraph } = useGraphFlow();

	const handleDialogOkay = () => {
		localStorage.removeItem("Sources");
		localStorage.removeItem("Destination");
		localStorage.removeItem("pipelinename");
		localStorage.removeItem("selectedAgentIds");
		localStorage.removeItem("changesLog");
		localStorage.removeItem("platform");

		resetGraph();

		setCurrentStep(0);
		setIsDialogOpen(false);
		setIsSheetOpen(false);
	};

	const handleDialogCancel = () => {
		setIsDialogOpen(false);
	};

	const getDataFromChild = (pipelineId: string, pipelineName: string) => {
		setPipelineId(pipelineId);
		setPipelineName(pipelineName);
	};

	return (
		<div className="flex flex-col gap-7 justify-center items-center">
			<Sheet
				open={isSheetOpen}
				onOpenChange={open => {
					if (!open) {
						setIsDialogOpen(true);
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
					{currentStep == 0 ? (
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
			<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
				<DialogContent className="w-[50rem]">
					<DialogHeader>
						<DialogTitle>Discard Changes?</DialogTitle>
						<DialogDescription>
							Are you sure you want to discard the current pipeline setup?
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
		</div>
	);
};

export default AddPipelineSheet;

import { Button } from "../../ui/button";
import { PlusIcon } from "lucide-react";
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from "@/components/ui/sheet";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogTitle,
	DialogDescription,
} from "@/components/ui/dialog";
import AddPipelineDetails from "./AddPipelineDetails";
import { useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import PipelineEditorSheet from "../editor/PipelineGraphEditor";
import KubernetesSetup from "../kubernetes/KubernetesSetup";
import { refreshKubernetesStatus } from "@/hooks/useKubernetesInstallations";

const AddPipelineSheet = () => {
	const [mode, setMode] = useState<"vm" | "kubernetes">();
	const [currentStep, setCurrentStep] = useState(0);
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [isDialogOpen, setIsDialogOpen] = useState(false);
	const [pipelineId, setPipelineId] = useState("");
	const [pipelineName, setPipelineName] = useState("");
	const [kubernetesCreated, setKubernetesCreated] = useState(false);
	const { resetGraph, changesLog } = useGraphFlow();

	const close = () => {
		if (mode === "vm")
			for (const key of [
				"Sources",
				"Destination",
				"pipelinename",
				"selectedAgentIds",
				"changesLog",
				"platform",
			])
				localStorage.removeItem(key);
		resetGraph();
		setCurrentStep(0);
		setMode(undefined);
		setKubernetesCreated(false);
		setPipelineId("");
		setPipelineName("");
		setIsSheetOpen(false);
		setIsDialogOpen(false);
		refreshKubernetesStatus();
	};
	const setOpen = (open: boolean) => {
		if (open) setIsSheetOpen(true);
		else if (changesLog.length || (mode && currentStep === 0) || kubernetesCreated)
			setIsDialogOpen(true);
		else close();
	};
	return (
		<>
			<Sheet open={isSheetOpen} onOpenChange={setOpen}>
				<SheetTrigger asChild>
					<Button className="flex gap-1 bg-blue-500 text-white" variant="outline">
						<PlusIcon className="h-4 w-4" />
						Add new pipeline
					</Button>
				</SheetTrigger>
				<SheetContent className={currentStep === 1 ? "w-screen p-0" : "overflow-y-auto"}>
					<SheetTitle className="sr-only">Create a pipeline</SheetTitle>
					{!mode ? (
						<div className="mx-auto max-w-2xl space-y-5 py-10">
							<h2 className="text-2xl font-semibold">Where will collectors run?</h2>
							<button
								onClick={() => setMode("kubernetes")}
								className="w-full rounded-lg border p-5 text-left hover:border-blue-500"
							>
								<span className="block text-lg font-medium">Kubernetes</span>
								<span className="mt-1 block text-sm text-gray-600">
									One collector per node, shared configuration, pod logs, and node/pod/container metrics.
								</span>
							</button>
							<button
								onClick={() => setMode("vm")}
								className="w-full rounded-lg border p-5 text-left hover:border-blue-500"
							>
								<span className="block text-lg font-medium">Linux or macOS</span>
								<span className="mt-1 block text-sm text-gray-600">
									Install a collector directly on a VM or host.
								</span>
							</button>
						</div>
					) : (
						<>
							{currentStep === 0 && !kubernetesCreated && (
								<Button variant="ghost" className="mb-4" onClick={() => setMode(undefined)}>
									Back to environments
								</Button>
							)}
							{mode === "kubernetes" && (
								<div hidden={currentStep !== 0}>
									<KubernetesSetup
										onCreated={() => setKubernetesCreated(true)}
										onConfigure={(id, name) => {
											setPipelineId(id);
											setPipelineName(name);
											setCurrentStep(1);
										}}
									/>
								</div>
							)}
							{mode === "vm" && currentStep === 0 && (
								<AddPipelineDetails
									currentStep={currentStep}
									setCurrentStep={setCurrentStep}
									sendPipelineDataToParent={(id, name) => {
										setPipelineId(id);
										setPipelineName(name);
									}}
								/>
							)}
							{currentStep === 1 && (
								<PipelineEditorSheet
									pipelineId={pipelineId}
									name={pipelineName}
									setIsSheetOpen={value => {
										const open = typeof value === "function" ? value(isSheetOpen) : value;
										if (!open && mode === "kubernetes") {
											setCurrentStep(0);
											refreshKubernetesStatus();
										} else if (!open) close();
									}}
									isEditModeStart
									isKubernetes={mode === "kubernetes"}
								/>
							)}
						</>
					)}
				</SheetContent>
			</Sheet>
			<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
				<DialogContent>
					<DialogTitle>
						{kubernetesCreated ? "Close Kubernetes setup?" : "Discard unsaved setup?"}
					</DialogTitle>
					<DialogDescription>
						{kubernetesCreated
							? "The installation and saved pipeline remain available. Keep a copy of the manifest or a download command; its installation token is available only during this setup session. Unsaved configuration edits will be lost."
							: "Unsaved form settings and configuration edits will be lost."}
					</DialogDescription>
					<DialogFooter>
						<Button variant="outline" onClick={() => setIsDialogOpen(false)}>
							Continue setup
						</Button>
						<Button onClick={close}>Close setup</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
};

export default AddPipelineSheet;

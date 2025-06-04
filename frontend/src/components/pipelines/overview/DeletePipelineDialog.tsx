import {
	Dialog,
	DialogTrigger,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
	DialogFooter,
	DialogClose,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";
import { Pipeline } from "@/types/pipeline.types";
import pipelineServices from "@/services/pipelineServices";
import { toast } from "@/hooks/use-toast";
import { useGraphFlow } from "@/context/useGraphFlowContext";

interface Props {
	isOpen: boolean;
	setIsOpen: (open: boolean) => void;
	pipelineOverview?: Pipeline;
}

const DeletePipelineDialog = ({ isOpen, setIsOpen, pipelineOverview }: Props) => {
	const { resetGraph } = useGraphFlow();
	const handleDeletePipeline = async () => {
		try {
			if (pipelineOverview?.id) {
				await pipelineServices.deletePipelineById(pipelineOverview.id);
			}

			setIsOpen(false);
			resetGraph();
			window.location.reload();
		} catch (error) {
			console.error("Error deleting pipeline or collector:", error);
			toast({
				title: "Error",
				description: "Failed to delete pipeline or collector",
				variant: "destructive",
			});
		}
	};

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger asChild>
				<Button variant="destructive">Delete Pipeline</Button>
			</DialogTrigger>

			<DialogContent className="sm:max-w-[38rem]">
				<DialogHeader>
					<DialogTitle className="text-red-600 text-2xl font-semibold">Delete Pipeline</DialogTitle>
					<DialogDescription className="text-base text-gray-700 mt-1">
						Are you sure you want to permanently delete this pipeline? This action cannot be undone.
					</DialogDescription>
				</DialogHeader>

				<div className="mt-4 space-y-2 text-sm text-gray-800">
					<p>
						<span className="font-medium">Pipeline ID:</span> {pipelineOverview?.id}
					</p>
					<p>
						<span className="font-medium">Pipeline Name:</span> {pipelineOverview?.name}
					</p>
				</div>

				<div
					role="alert"
					className="mt-4 flex items-start gap-3 bg-yellow-50 border border-yellow-300 text-yellow-800 px-4 py-3 rounded-lg shadow-sm">
					<div className="pt-1">
						<div className="bg-yellow-100 rounded-full p-1">
							<AlertTriangle className="w-5 h-5 text-yellow-600" />
						</div>
					</div>
					<p className="text-sm leading-snug">
						<strong>Warning:</strong> Stopping this pipeline will stop the collector on the associated
						VM/Kubernetes node.
					</p>
				</div>

				<DialogFooter className="mt-6">
					<DialogClose asChild>
						<Button variant="outline">Cancel</Button>
					</DialogClose>
					<Button onClick={handleDeletePipeline} variant="destructive">
						Delete
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
};

export default DeletePipelineDialog;

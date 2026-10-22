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
import { useEffect, useRef, useState, type ReactNode } from "react";
import pipelineServices from "@/services/pipeline";
import kubernetesService from "@/services/kubernetes";
import { toast } from "@/hooks/useToast";
import { refreshKubernetesStatus } from "@/hooks/useKubernetesInstallations";
import type { KubernetesInstallation } from "@/types/kubernetes.types";
import { installationCommands } from "@/utils/kubernetes";
import { ManifestCodeBlock } from "../kubernetes/ManifestCodeBlock";

interface Props {
	pipelineId: string;
	pipelineName: string;
	onDeleted: () => void;
	children?: ReactNode;
}

const DeletePipelineDialog = ({ pipelineId, pipelineName, onDeleted, children }: Props) => {
	const [isOpen, setIsOpen] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);
	const deleting = useRef(false);
	const cancelButton = useRef<HTMLButtonElement>(null);
	const [installations, setInstallations] = useState<KubernetesInstallation[] | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState("");
	const [loadAttempt, setLoadAttempt] = useState(0);

	useEffect(() => {
		if (!isOpen) return;
		let active = true;
		const controller = new AbortController();
		setLoading(true);
		setInstallations(null);
		setError("");
		kubernetesService
			.getPipelineInstallations(pipelineId, controller.signal)
			.then(data => {
				if (active) setInstallations(data ?? []);
			})
			.catch((error: unknown) => {
				if (active)
					setError(error instanceof Error ? error.message : "Could not load pipeline installations.");
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
			controller.abort();
		};
	}, [isOpen, pipelineId, loadAttempt]);

	const handleDeletePipeline = async () => {
		if (deleting.current || loading || !installations) return;
		deleting.current = true;
		setIsDeleting(true);
		setError("");
		try {
			await pipelineServices.deletePipelineById(pipelineId);
		} catch (error: unknown) {
			setError(error instanceof Error ? error.message : "Failed to delete pipeline. Try again.");
			return;
		} finally {
			deleting.current = false;
			setIsDeleting(false);
		}
		setIsOpen(false);
		onDeleted();
		refreshKubernetesStatus();
		toast({ title: "Pipeline deleted", description: `${pipelineName} was deleted.` });
	};

	return (
		<Dialog
			open={isOpen}
			onOpenChange={open => {
				if (!deleting.current) setIsOpen(open);
			}}
		>
			<DialogTrigger asChild>
				{children ?? <Button variant="destructive">Delete Pipeline</Button>}
			</DialogTrigger>

			<DialogContent
				className="max-h-[90vh] overflow-y-auto sm:max-w-[38rem]"
				aria-busy={isDeleting}
				onOpenAutoFocus={event => {
					event.preventDefault();
					cancelButton.current?.focus();
				}}
			>
				<DialogHeader>
					<DialogTitle className="text-red-600 text-2xl font-semibold">Delete Pipeline</DialogTitle>
					<DialogDescription className="text-base text-gray-700 mt-1">
						Permanently delete {pipelineName}? This action cannot be undone.
					</DialogDescription>
				</DialogHeader>

				<div className="mt-4 space-y-2 text-sm text-gray-800">
					<p>
						<span className="font-medium">Pipeline ID:</span> {pipelineId}
					</p>
					<p>
						<span className="font-medium">Pipeline Name:</span> {pipelineName}
					</p>
				</div>

				<div
					role="alert"
					className="mt-4 flex items-start gap-3 bg-yellow-50 border border-yellow-300 text-yellow-800 px-4 py-3 rounded-lg shadow-sm"
				>
					<div className="pt-1">
						<div className="bg-yellow-100 rounded-full p-1">
							<AlertTriangle className="w-5 h-5 text-yellow-600" />
						</div>
					</div>
					<p className="text-sm leading-snug">
						The pipeline configuration will be removed and VM collectors will be detached. Associated
						Kubernetes installations and collector records will be deleted, and their credentials revoked.
						Running collectors may continue using their last configuration.
					</p>
				</div>

				{loading && (
					<p role="status" className="text-sm text-gray-500">
						Checking associated installations...
					</p>
				)}
				{!!installations?.length && (
					<div className="min-w-0 space-y-3">
						<p className="text-sm text-gray-700">
							To stop Kubernetes collectors, run these commands in each installation's cluster before
							deleting. The observability namespace is preserved. Save the commands if you need them later.
						</p>
						{installations.map(installation => (
							<ManifestCodeBlock
								key={installation.id}
								title={`Uninstall ${installation.name} in ${installation.cluster_name}`}
								code={installationCommands(installation.id).uninstall}
							/>
						))}
					</div>
				)}
				{error && (
					<p role="alert" className="text-sm text-red-700">
						{error}
					</p>
				)}
				{!loading && !installations && (
					<Button variant="outline" onClick={() => setLoadAttempt(attempt => attempt + 1)}>
						Retry loading installations
					</Button>
				)}
				<DialogFooter className="mt-6">
					<DialogClose asChild>
						<Button ref={cancelButton} variant="outline" disabled={isDeleting}>
							Cancel
						</Button>
					</DialogClose>
					<Button
						onClick={handleDeletePipeline}
						variant="destructive"
						disabled={isDeleting || loading || !installations}
					>
						{isDeleting ? "Deleting..." : "Delete"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
};

export default DeletePipelineDialog;

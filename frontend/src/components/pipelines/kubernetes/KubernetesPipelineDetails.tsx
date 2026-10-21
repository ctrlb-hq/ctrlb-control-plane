import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogTitle,
} from "@/components/ui/dialog";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import {
	refreshKubernetesStatus,
	useKubernetesInstallations,
} from "@/hooks/useKubernetesInstallations";
import { installationCommands } from "@/utils/kubernetes";
import PipelineEditorSheet from "../editor/PipelineGraphEditor";
import { InstallationStatus } from "./InstallationStatus";

export default function KubernetesPipelineDetails({
	pipelineId,
	name,
}: {
	pipelineId: string;
	name: string;
}) {
	const { installations, error, loading, lastUpdated } = useKubernetesInstallations({ pipelineId });
	const [editing, setEditing] = useState(false);
	const [discarding, setDiscarding] = useState(false);
	const { changesLog, clearChangesLog } = useGraphFlow();
	const setEditorOpen = (open: boolean) => {
		if (!open && changesLog.length) setDiscarding(true);
		else setEditing(open);
	};
	return (
		<div className="flex h-full flex-col gap-5 overflow-y-auto p-2">
			<div className="flex flex-wrap items-start justify-between gap-3 pr-8">
				<div>
					<h2 className="text-2xl font-semibold">{name}</h2>
					<p className="mt-1 text-sm text-gray-600">
						Kubernetes collector installations · Pipeline {pipelineId}
					</p>
				</div>
				<div className="flex gap-2">
					<Button variant="outline" onClick={refreshKubernetesStatus}>
						Refresh
					</Button>
					<Button onClick={() => setEditing(true)}>View/edit configuration</Button>
				</div>
			</div>
			<p className="text-sm text-gray-600">
				Each collector applies changes independently and keeps its last working configuration if an
				update fails. Engine state reports the collector; confirm exported telemetry in CtrlB.
			</p>
			{error && (
				<p role="alert" className="rounded bg-amber-50 p-3 text-sm text-amber-800">
					{error}
				</p>
			)}
			{loading && !installations.length && (
				<p role="status" className="text-sm text-gray-500">
					Loading collector status...
				</p>
			)}
			{!loading && !error && !installations.length && (
				<p className="text-sm text-gray-500">
					No Kubernetes installations are assigned to this pipeline.
				</p>
			)}
			{installations.map(installation => {
				const commands = installationCommands(installation.id);
				return (
					<div key={installation.id} className="space-y-2">
						<InstallationStatus installation={installation} />
						<details className="rounded border p-3 text-sm">
							<summary className="cursor-pointer text-gray-600">
								Kubernetes rollout and uninstall commands
							</summary>
							<pre className="mt-3 overflow-auto rounded bg-gray-50 p-3">{commands.rollout}</pre>
							<p className="mt-3 text-gray-600">
								Uninstall only this installation's resources. The shared observability namespace is
								preserved.
							</p>
							<pre className="mt-2 overflow-auto rounded bg-gray-50 p-3">{commands.uninstall}</pre>
						</details>
					</div>
				);
			})}
			<p className="text-xs text-gray-500">
				Refreshes every 10 seconds.{" "}
				{lastUpdated
					? `Last refreshed ${new Date(lastUpdated).toLocaleString()}.`
					: "Waiting for status."}
			</p>
			<Sheet open={editing} onOpenChange={setEditorOpen}>
				<SheetContent className="w-screen p-0">
					<SheetTitle className="sr-only">Edit Kubernetes pipeline</SheetTitle>
					<PipelineEditorSheet
						pipelineId={pipelineId}
						name={name}
						setIsSheetOpen={setEditing}
						isEditModeStart={false}
						isKubernetes
					/>
				</SheetContent>
			</Sheet>
			<Dialog open={discarding} onOpenChange={setDiscarding}>
				<DialogContent>
					<DialogTitle>Discard configuration edits?</DialogTitle>
					<DialogDescription>
						Your saved configuration remains in place. Unsaved edits will be lost.
					</DialogDescription>
					<DialogFooter>
						<Button variant="outline" onClick={() => setDiscarding(false)}>
							Keep editing
						</Button>
						<Button
							onClick={() => {
								clearChangesLog();
								setDiscarding(false);
								setEditing(false);
							}}
						>
							Discard edits
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}

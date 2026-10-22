import { Sheet, SheetContent, SheetTrigger, SheetTitle } from "@/components/ui/sheet";
import {
	Table,
	TableBody,
	TableCaption,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import pipelineServices from "@/services/pipeline";
import kubernetesService from "@/services/kubernetes";
import { useEffect, useState } from "react";
import type { KubernetesInstallation } from "@/types/kubernetes.types";
import { summarizeInstallations } from "@/utils/kubernetes";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";
import ViewPipelineDetails from "./ViewPipelineDetails";
import KubernetesPipelineDetails from "../kubernetes/KubernetesPipelineDetails";
import DeletePipelineDialog from "./DeletePipelineDialog";

interface PipelineRow {
	id: string | number;
	name: string;
	agents: number;
	incoming_bytes: number;
	outgoing_bytes: number;
	updatedAt: number;
}

const PipelineTable = () => {
	const [pipelines, setPipelines] = useState<PipelineRow[]>([]);
	const [installations, setInstallations] = useState<KubernetesInstallation[]>([]);
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(true);
	const [openPipeline, setOpenPipeline] = useState<string>();
	const { resetGraph } = useGraphFlow();
	const handleDeleted = (id: string) => {
		setPipelines(current => current.filter(pipeline => String(pipeline.id) !== id));
		setInstallations(current =>
			current.filter(installation => String(installation.pipeline_id) !== id),
		);
		if (openPipeline === id) {
			setOpenPipeline(undefined);
			resetGraph();
		}
	};
	useEffect(() => {
		let active = true;
		let pending = false;
		let refreshRequested = false;
		const refresh = async (invalidatePending = false) => {
			if (pending) {
				if (invalidatePending) refreshRequested = true;
				return;
			}
			pending = true;
			try {
				const [pipelines, installations] = await Promise.all([
					pipelineServices.getAllPipelines(),
					kubernetesService.listInstallations(),
				]);
				if (active && !refreshRequested) {
					setPipelines(pipelines ?? []);
					setInstallations(installations ?? []);
					setError("");
				}
			} catch {
				if (active && !refreshRequested)
					setError(
						"Could not refresh pipeline status. Retrying every 10 seconds; shown data may be stale.",
					);
			} finally {
				pending = false;
				if (active && refreshRequested) {
					refreshRequested = false;
					void refresh();
				} else if (active) setLoading(false);
			}
		};
		const handleUpdated = () => void refresh(true);
		void refresh();
		const timer = window.setInterval(() => void refresh(), 10_000);
		window.addEventListener("pipelines-updated", handleUpdated);
		return () => {
			active = false;
			window.clearInterval(timer);
			window.removeEventListener("pipelines-updated", handleUpdated);
		};
	}, []);
	return (
		<div className="space-y-3">
			{error && (
				<p role="alert" className="text-sm text-amber-800">
					{error}
				</p>
			)}
			{loading && (
				<p role="status" className="text-sm text-gray-500">
					Loading pipelines...
				</p>
			)}
			{!loading && !pipelines.length && !error && (
				<div className="space-y-2 py-12 text-center">
					<h2 className="text-xl font-semibold">Create your first pipeline</h2>
					<p className="text-gray-600">
						Install Kubernetes collectors or connect a VM collector to start routing telemetry.
					</p>
				</div>
			)}
			{!!pipelines.length && (
				<Table className="border border-gray-200">
					<TableCaption>Pipeline and collector status refreshes every 10 seconds.</TableCaption>
					<TableHeader className="bg-gray-100">
						<TableRow>
							<TableHead>Name</TableHead>
							<TableHead>Environment</TableHead>
							<TableHead>Collectors</TableHead>
							<TableHead>Configuration rollout</TableHead>
							<TableHead>Traffic</TableHead>
							<TableHead>Updated</TableHead>
							<TableHead className="text-right">Actions</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{pipelines.map(pipeline => {
							const id = String(pipeline.id);
							const fleet = installations.filter(installation => String(installation.pipeline_id) === id);
							const counts = summarizeInstallations(fleet);
							return (
								<Sheet
									key={id}
									open={openPipeline === id}
									onOpenChange={open => {
										setOpenPipeline(open ? id : undefined);
										if (!open) resetGraph();
									}}
								>
									<TableRow className="cursor-pointer" onClick={() => setOpenPipeline(id)}>
										<TableCell className="font-medium">
											<SheetTrigger asChild>
												<Button variant="link" className="h-auto p-0 text-inherit">
													{pipeline.name}
												</Button>
											</SheetTrigger>
										</TableCell>
										<TableCell>{fleet.length ? "Kubernetes" : "VM"}</TableCell>
										<TableCell>
											{fleet.length
												? counts.total
													? `${counts.connected}/${counts.total} connected`
													: "Waiting for enrollment"
												: `${pipeline.agents} enrolled`}
										</TableCell>
										<TableCell>
											{fleet.length ? (
												<div className="space-y-1 text-xs">
													<p>
														{counts.applied} applied · {counts.pending} pending · {counts.failed} failed
													</p>
													{counts.disconnected > 0 && (
														<p className="text-red-700">{counts.disconnected} disconnected</p>
													)}
												</div>
											) : (
												"Direct configuration"
											)}
										</TableCell>
										<TableCell>
											{fleet.length
												? "See collector metrics"
												: `${pipeline.incoming_bytes} B in / ${pipeline.outgoing_bytes} B out`}
										</TableCell>
										<TableCell className="text-xs text-gray-500">
											{pipeline.updatedAt
												? new Date(pipeline.updatedAt * 1000).toLocaleString()
												: "Unavailable"}
										</TableCell>
										<TableCell className="text-right" onClick={event => event.stopPropagation()}>
											<DeletePipelineDialog
												pipelineId={id}
												pipelineName={pipeline.name}
												onDeleted={() => handleDeleted(id)}
											>
												<Button
													variant="ghost"
													size="sm"
													className="text-red-700 hover:text-red-800"
													aria-label={`Delete pipeline ${pipeline.name}`}
												>
													<Trash2 className="mr-2 h-4 w-4" aria-hidden="true" />
													Delete
												</Button>
											</DeletePipelineDialog>
										</TableCell>
									</TableRow>
									<SheetContent className="overflow-y-auto">
										<SheetTitle className="sr-only">{pipeline.name}</SheetTitle>
										{openPipeline === id &&
											(fleet.length ? (
												<KubernetesPipelineDetails
													pipelineId={id}
													name={pipeline.name}
													onDeleted={() => handleDeleted(id)}
												/>
											) : (
												<ViewPipelineDetails
													pipelineId={id}
													name={pipeline.name}
													onDeleted={() => handleDeleted(id)}
												/>
											))}
									</SheetContent>
								</Sheet>
							);
						})}
					</TableBody>
				</Table>
			)}
		</div>
	);
};

export default PipelineTable;

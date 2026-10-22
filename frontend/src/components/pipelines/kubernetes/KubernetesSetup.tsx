import { useEffect, useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import kubernetesService from "@/services/kubernetes";
import {
	installationCommands,
	kubernetesDefaults,
	manifestDownloadCommands,
	validateHTTPURL,
} from "@/utils/kubernetes";
import { API_BASE_URL } from "@/config";
import {
	refreshKubernetesStatus,
	useKubernetesInstallations,
} from "@/hooks/useKubernetesInstallations";
import type { InstallationCreation } from "@/types/kubernetes.types";
import { InstallationStatus } from "./InstallationStatus";
import { ManifestCodeBlock } from "./ManifestCodeBlock";

export default function KubernetesSetup({
	onConfigure,
	onCreated,
}: {
	onConfigure: (pipelineId: string, name: string) => void;
	onCreated: () => void;
}) {
	const [name, setName] = useState("");
	const [clusterName, setClusterName] = useState("");
	const [backendURL, setBackendURL] = useState(API_BASE_URL);
	const [endpoint, setEndpoint] = useState(kubernetesDefaults.endpoint);
	const [logsEndpoint, setLogsEndpoint] = useState(kubernetesDefaults.logsEndpoint);
	const [metricsEndpoint, setMetricsEndpoint] = useState(kubernetesDefaults.metricsEndpoint);
	const [authorization, setAuthorization] = useState("");
	const [image, setImage] = useState(kubernetesDefaults.image);
	const [skipVerify, setSkipVerify] = useState(false);
	const [pipelineId, setPipelineId] = useState<number>();
	const [creation, setCreation] = useState<InstallationCreation>();
	const [busy, setBusy] = useState(false);
	const [manifest, setManifest] = useState<{ yaml: string; backendURL: string; image: string }>();
	const [manifestSaved, setManifestSaved] = useState(false);
	const [manifestAttempt, setManifestAttempt] = useState(0);
	const [error, setError] = useState("");
	const [downloadError, setDownloadError] = useState("");
	const status = useKubernetesInstallations({
		installationId: creation?.installation.id,
		enabled: !!creation,
	});
	const current = status.installations[0] ?? creation?.installation;
	const currentManifest =
		manifest?.backendURL === backendURL && manifest?.image === image ? manifest.yaml : undefined;

	useEffect(() => {
		if (!creation) return;
		const controller = new AbortController();
		setManifest(undefined);
		setManifestSaved(false);
		setDownloadError("");
		const timer = window.setTimeout(() => {
			void kubernetesService
				.manifest(creation.installation.id, creation.token, backendURL, image, controller.signal)
				.then(yaml => {
					if (!controller.signal.aborted) setManifest({ yaml, backendURL, image });
				})
				.catch((error: unknown) => {
					if (!controller.signal.aborted)
						setDownloadError(error instanceof Error ? error.message : "Could not generate the manifest.");
				});
		}, 300);
		return () => {
			window.clearTimeout(timer);
			controller.abort();
		};
	}, [creation, backendURL, image, manifestAttempt]);

	const create = async (event: FormEvent) => {
		event.preventDefault();
		setError("");
		setBusy(true);
		try {
			if (!name.trim() || !clusterName.trim())
				throw new Error("Enter a pipeline name and cluster name.");
			validateHTTPURL(backendURL, "Control-plane URL");
			validateHTTPURL(endpoint, "Telemetry base endpoint");
			if (logsEndpoint) validateHTTPURL(logsEndpoint, "Logs endpoint");
			if (metricsEndpoint) validateHTTPURL(metricsEndpoint, "Metrics endpoint");
			if (!image.trim() || /\s/.test(image.trim()))
				throw new Error("Enter a collector image reference without spaces.");
			let id = pipelineId;
			if (!id) {
				const pipeline = await kubernetesService.createPipeline({
					name: name.trim(),
					endpoint,
					logs_endpoint: logsEndpoint || undefined,
					metrics_endpoint: metricsEndpoint || undefined,
					headers: authorization ? { Authorization: authorization } : {},
					kubelet_insecure_skip_verify: skipVerify,
				});
				id = pipeline.pipeline_id;
				setPipelineId(id);
				setAuthorization("");
			}
			const result = await kubernetesService.createInstallation({
				name: name.trim(),
				cluster_name: clusterName.trim(),
				pipeline_id: id,
			});
			setCreation(result);
			onCreated();
			refreshKubernetesStatus();
		} catch (error) {
			setError(error instanceof Error ? error.message : "Could not create the installation.");
		} finally {
			setBusy(false);
		}
	};

	const download = () => {
		if (!creation || !currentManifest) return;
		setDownloadError("");
		try {
			const commands = installationCommands(creation.installation.id);
			const url = URL.createObjectURL(new Blob([currentManifest], { type: "application/yaml" }));
			const link = document.createElement("a");
			link.href = url;
			link.download = commands.filename;
			document.body.appendChild(link);
			link.click();
			link.remove();
			window.setTimeout(() => URL.revokeObjectURL(url), 1_000);
			setManifestSaved(true);
		} catch (error) {
			setDownloadError(error instanceof Error ? error.message : "Could not download the manifest.");
		}
	};

	if (creation && current) {
		const commands = installationCommands(creation.installation.id);
		const downloads = currentManifest
			? manifestDownloadCommands(
					creation.installation.id,
					creation.token,
					backendURL,
					image,
					API_BASE_URL,
				)
			: undefined;
		return (
			<div className="space-y-6">
				<div>
					<h2 className="text-2xl font-semibold">Install collectors in {current.cluster_name}</h2>
					<p className="mt-1 text-sm text-gray-600">
						The shared pipeline is saved. Copy or download the manifest, then apply it to your cluster.
					</p>
				</div>
				<div className="space-y-4 rounded-lg border p-4">
					<div className="grid gap-4 md:grid-cols-2">
						<div className="space-y-1">
							<Label htmlFor="manifest-backend-url">Collector-reachable control-plane URL</Label>
							<Input
								id="manifest-backend-url"
								type="url"
								value={backendURL}
								onChange={event => setBackendURL(event.target.value)}
							/>
						</div>
						<div className="space-y-1">
							<Label htmlFor="manifest-image">Collector image</Label>
							<Input id="manifest-image" value={image} onChange={event => setImage(event.target.value)} />
						</div>
					</div>
					<p className="text-sm text-gray-600">
						The YAML and download commands include the installation token. Keep a private copy before
						closing setup; the token is available only during this setup session.
					</p>
					<Button
						className="h-auto max-w-full whitespace-normal"
						onClick={download}
						disabled={!currentManifest}
					>
						Download installation manifest
					</Button>
					{!currentManifest && !downloadError && (
						<p role="status" className="text-sm text-gray-600">
							Generating manifest...
						</p>
					)}
					{downloadError && (
						<div className="space-y-2">
							<p role="alert" className="text-sm text-red-700">
								{downloadError}
							</p>
							<Button variant="outline" onClick={() => setManifestAttempt(attempt => attempt + 1)}>
								Retry manifest generation
							</Button>
						</div>
					)}
					{currentManifest && downloads && (
						<div className="space-y-3 text-sm">
							<ManifestCodeBlock
								title="Manifest YAML"
								code={currentManifest}
								onCopy={() => setManifestSaved(true)}
							/>
							<p>Run either download command on the machine you use to control your cluster.</p>
							<ManifestCodeBlock
								title="curl"
								code={downloads.curl}
								onCopy={() => setManifestSaved(true)}
							/>
							<ManifestCodeBlock
								title="wget"
								code={downloads.wget}
								onCopy={() => setManifestSaved(true)}
							/>
							<p>Save the YAML as {commands.filename}, or use a download command above. Then run:</p>
							<ManifestCodeBlock
								title="Apply and check rollout"
								code={`${commands.apply}\n${commands.rollout}`}
							/>
							<details>
								<summary className="cursor-pointer text-gray-600">Uninstall this installation</summary>
								<pre className="mt-2 overflow-auto rounded bg-gray-100 p-3">{commands.uninstall}</pre>
								<p className="mt-1 text-xs text-gray-500">
									The shared observability namespace and other installations are preserved.
								</p>
							</details>
						</div>
					)}
				</div>
				{status.error && (
					<p role="alert" className="text-sm text-amber-800">
						{status.error}
					</p>
				)}
				<InstallationStatus installation={current} />
				<div className="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
					<p className="text-xs text-gray-500">
						Collector status refreshes every 10 seconds.{" "}
						{!manifestSaved &&
							"Copy the YAML or a download command, or download the file, before opening the editor."}
					</p>
					<Button
						className="h-auto max-w-full whitespace-normal"
						variant="outline"
						disabled={!manifestSaved || !currentManifest}
						onClick={() => onConfigure(String(current.pipeline_id), name)}
					>
						View/edit pipeline configuration
					</Button>
				</div>
			</div>
		);
	}

	return (
		<form onSubmit={event => void create(event)} autoComplete="off" className="space-y-6">
			<div>
				<h2 className="text-2xl font-semibold">Kubernetes collectors</h2>
				<p className="mt-1 text-sm text-gray-600">
					Collect pod logs from all namespaces and node, pod, and container metrics every 30 seconds.
				</p>
			</div>
			<fieldset disabled={busy || !!pipelineId} className="grid gap-4 md:grid-cols-2">
				<div className="space-y-1">
					<Label htmlFor="k8s-name">Pipeline / installation name</Label>
					<Input
						id="k8s-name"
						value={name}
						onChange={event => setName(event.target.value)}
						required
						maxLength={253}
					/>
				</div>
				<div className="space-y-1">
					<Label htmlFor="k8s-cluster">Cluster name</Label>
					<Input
						id="k8s-cluster"
						value={clusterName}
						onChange={event => setClusterName(event.target.value)}
						required
						maxLength={253}
					/>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-backend">Collector-reachable control-plane URL</Label>
					<Input
						id="k8s-backend"
						type="url"
						placeholder="https://ctrlb.explore.ctrlb.ai/control-plane/api"
						value={backendURL}
						onChange={event => setBackendURL(event.target.value)}
						required
					/>
					<p className="text-xs text-gray-500">
						Every node must reach this management URL. It is separate from the telemetry destination
						below. Use HTTPS, or HTTP for a local demo.
					</p>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-image">Collector image</Label>
					<Input
						id="k8s-image"
						value={image}
						onChange={event => setImage(event.target.value)}
						required
					/>
					<p className="text-xs text-gray-500">
						Build and load ctrlb-agent:k8s-v1 onto the cluster nodes, or enter an image your cluster can
						pull.
					</p>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-endpoint">CtrlB OTLP HTTP base endpoint</Label>
					<Input
						id="k8s-endpoint"
						type="url"
						value={endpoint}
						onChange={event => setEndpoint(event.target.value)}
						required
					/>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-logs">Logs OTLP endpoint</Label>
					<Input
						id="k8s-logs"
						type="url"
						value={logsEndpoint}
						onChange={event => setLogsEndpoint(event.target.value)}
					/>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-metrics">Metrics OTLP endpoint</Label>
					<Input
						id="k8s-metrics"
						type="url"
						value={metricsEndpoint}
						onChange={event => setMetricsEndpoint(event.target.value)}
					/>
				</div>
				<div className="space-y-1 md:col-span-2">
					<Label htmlFor="k8s-auth">CtrlB Authorization header value</Label>
					<Input
						id="k8s-auth"
						type="password"
						autoComplete="off"
						spellCheck={false}
						placeholder="Enter the required authorization value"
						value={authorization}
						onChange={event => setAuthorization(event.target.value)}
					/>
					<p className="text-xs text-gray-500">
						Saved with the pipeline on the control plane. It is never saved in this browser's local
						storage.
					</p>
				</div>
				<label className="flex items-start gap-2 text-sm md:col-span-2">
					<input
						type="checkbox"
						className="mt-1"
						checked={skipVerify}
						onChange={event => setSkipVerify(event.target.checked)}
					/>
					<span>
						Skip kubelet certificate verification for this installation.
						<span className="block text-xs text-gray-500">
							For development clusters such as kind with a self-signed kubelet certificate. Verification is
							enabled by default.
						</span>
					</span>
				</label>
			</fieldset>
			{pipelineId && (
				<p className="text-sm text-gray-600">
					Pipeline {pipelineId} is saved. Retry to create its installation.
				</p>
			)}
			{error && (
				<p role="alert" className="text-sm text-red-700">
					{error}
				</p>
			)}
			<Button type="submit" disabled={busy}>
				{busy
					? "Saving setup..."
					: pipelineId
						? "Retry installation creation"
						: "Create pipeline and installation"}
			</Button>
		</form>
	);
}

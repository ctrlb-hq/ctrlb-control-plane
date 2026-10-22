import type {
	KubernetesCollector,
	KubernetesInstallation,
	RolloutSummary,
} from "../types/kubernetes.types.ts";

export const kubernetesDefaults = {
	endpoint: "https://ctrlb.explore.ctrlb.ai/engine/api/default",
	logsEndpoint: "https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs",
	metricsEndpoint: "https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics",
	image: "ctrlb-agent:k8s-v1",
};

export function summarizeInstallations(installations: KubernetesInstallation[]) {
	const summary: RolloutSummary & { connected: number } = {
		total: 0,
		connected: 0,
		disconnected: 0,
		pending: 0,
		applied: 0,
		failed: 0,
	};
	for (const installation of installations) {
		for (const collector of installation.collectors) {
			summary.total++;
			if (collector.connection_status === "connected") summary.connected++;
			else summary.disconnected++;
			if (collector.rollout_status === "applied") summary.applied++;
			else if (collector.rollout_status === "failed") summary.failed++;
			else summary.pending++;
		}
	}
	return summary;
}

export function formatMetric(value: number | null | undefined, unit: string, divisor = 1) {
	if (value == null || !Number.isFinite(value) || value < 0) return "Unavailable";
	return `${(value / divisor).toLocaleString("en-US", { maximumFractionDigits: 1 })} ${unit}`;
}

export function shortRevision(value: string) {
	return value ? value.replace(/^sha256:/, "").slice(0, 12) : "Unavailable";
}

export function isPreviousConfigurationError(
	collector: KubernetesCollector,
	desiredRevision: string,
) {
	return Boolean(
		collector.config_error &&
			collector.engine_status === "running" &&
			collector.rollout_status === "applied" &&
			collector.applied_revision === desiredRevision &&
			collector.attempted_revision &&
			collector.attempted_revision !== desiredRevision,
	);
}

export function validateHTTPURL(value: string, label: string) {
	try {
		const url = new URL(value);
		if (
			!["http:", "https:"].includes(url.protocol) ||
			!url.hostname ||
			url.username ||
			url.password ||
			url.hash ||
			url.search
		)
			throw new Error();
		return value.replace(/\/+$/, "");
	} catch {
		throw new Error(
			`${label} must be an absolute HTTP or HTTPS URL without credentials, query, or fragment.`,
		);
	}
}

export function installationCommands(id: string) {
	if (!/^[a-f0-9]{32}$/.test(id)) throw new Error("Invalid installation ID.");
	const resource = `ctrlb-collector-${id.slice(0, 12)}`;
	const filename = `${resource}.yaml`;
	return {
		resource,
		filename,
		apply: `kubectl apply -f ${filename}`,
		rollout: `kubectl -n observability rollout status daemonset/${resource}`,
		uninstall: `kubectl -n observability delete daemonset,secret,serviceaccount ${resource}\nkubectl delete clusterrole,clusterrolebinding ${resource}`,
	};
}

export function buildManifestRequest(id: string, token: string, backendURL: string, image: string) {
	installationCommands(id);
	if (!token)
		throw new Error(
			"The installation token is unavailable. Keep the manifest downloaded during setup.",
		);
	if (!image.trim() || /\s/.test(image.trim()))
		throw new Error("Enter a collector image reference without spaces.");
	return {
		method: "POST" as const,
		url: `/installations/${encodeURIComponent(id)}/manifest`,
		data: {
			token,
			backend_url: validateHTTPURL(backendURL, "Control-plane URL"),
			image: image.trim(),
		},
		responseType: "text" as const,
	};
}

// Quote each argument for POSIX shells, including URLs containing apostrophes or shell syntax.
function shellQuote(value: string) {
	return `'${value.replace(/'/g, "'\\''")}'`;
}

export function manifestDownloadCommands(
	id: string,
	token: string,
	backendURL: string,
	image: string,
	apiBaseURL: string,
) {
	const request = buildManifestRequest(id, token, backendURL, image);
	const url = `${validateHTTPURL(apiBaseURL, "API URL")}/agent/v2/installations/${id}/manifest`;
	const body = shellQuote(JSON.stringify(request.data));
	const filename = shellQuote(installationCommands(id).filename);
	return {
		curl: [
			"curl --fail --silent --show-error",
			"  --header 'Content-Type: application/json'",
			`  --data ${body}`,
			`  --output ${filename}`,
			`  ${shellQuote(url)}`,
		].join(" \\\n"),
		wget: [
			"wget",
			"  --header='Content-Type: application/json'",
			`  --post-data=${body}`,
			`  --output-document=${filename}`,
			`  ${shellQuote(url)}`,
		].join(" \\\n"),
	};
}

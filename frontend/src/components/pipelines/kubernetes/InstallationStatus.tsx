import type { KubernetesInstallation } from "@/types/kubernetes.types";
import {
	formatMetric,
	isPreviousConfigurationError,
	shortRevision,
	summarizeInstallations,
} from "@/utils/kubernetes";

const time = (seconds?: number | null) =>
	seconds ? new Date(seconds * 1000).toLocaleString() : "Unavailable";
const stateStyle = (value: string) =>
	value === "failed" || value === "disconnected"
		? "bg-red-50 text-red-700"
		: value === "applied" || value === "running" || value === "connected"
			? "bg-green-50 text-green-700"
			: "bg-amber-50 text-amber-800";

function State({ value }: { value: string }) {
	return (
		<span className={`inline-block rounded px-2 py-1 text-xs font-medium ${stateStyle(value)}`}>
			{value}
		</span>
	);
}

export function InstallationStatus({ installation }: { installation: KubernetesInstallation }) {
	const counts = summarizeInstallations([installation]);
	return (
		<section
			className="rounded-lg border border-gray-200 bg-white"
			aria-label={`Installation ${installation.name}`}
		>
			<div className="space-y-2 border-b p-4">
				<div className="flex flex-wrap items-center justify-between gap-2">
					<h3 className="font-semibold">
						{installation.name}{" "}
						<span className="font-normal text-gray-500">/ {installation.cluster_name}</span>
					</h3>
					<p className="text-sm">
						{counts.connected} of {counts.total} enrolled collectors connected
					</p>
				</div>
				<p className="text-sm text-gray-600">
					{counts.applied} applied · {counts.pending} pending · {counts.failed} failed ·{" "}
					{counts.disconnected} disconnected
				</p>
				<p className="break-all text-xs text-gray-500">
					Installation {installation.id} · Desired revision{" "}
					<code title={installation.desired_revision}>
						{shortRevision(installation.desired_revision)}
					</code>
				</p>
			</div>
			{counts.total === 0 ? (
				<div className="space-y-1 p-6 text-sm text-gray-600" role="status">
					<p className="font-medium text-gray-800">Waiting for collectors to enroll</p>
					<p>
						Apply the installation manifest, then check the DaemonSet rollout. Nodes appear here after
						their first check-in.
					</p>
				</div>
			) : (
				<div className="overflow-x-auto">
					<table className="w-full min-w-[850px] text-left text-sm">
						<thead className="bg-gray-50 text-xs text-gray-600">
							<tr>
								<th className="p-3">Node / pod</th>
								<th className="p-3">Connection</th>
								<th className="p-3">Engine</th>
								<th className="p-3">Configuration</th>
								<th className="p-3">Latest collector metrics</th>
							</tr>
						</thead>
						<tbody>
							{installation.collectors.map(collector => (
								<tr key={collector.id} className="border-t align-top">
									<td className="max-w-[240px] space-y-2 p-3">
										<p className="break-all font-medium">{collector.node_name}</p>
										<p className="break-all text-xs text-gray-500">
											{collector.namespace}/{collector.pod_name || "Unavailable"}
										</p>
										<p className="text-xs text-gray-500">
											{collector.version || "Version unavailable"} · {collector.pod_ip || "IP unavailable"}
										</p>
										<details className="text-xs">
											<summary className="cursor-pointer text-blue-700">Identity</summary>
											<p className="mt-1 break-all">
												Collector {collector.id}
												<br />
												Pod UID {collector.pod_uid || "Unavailable"}
											</p>
										</details>
									</td>
									<td className="space-y-2 p-3">
										<State value={collector.connection_status} />
										<p className="text-xs text-gray-500">
											Last seen
											<br />
											{time(collector.last_seen)}
										</p>
									</td>
									<td className="p-3">
										<State value={collector.engine_status} />
									</td>
									<td className="max-w-[320px] space-y-2 p-3">
										<State value={collector.rollout_status} />
										<details className="text-xs">
											<summary className="cursor-pointer text-blue-700">Configuration revisions</summary>
											<dl className="mt-2 space-y-2">
												{[
													["Desired", installation.desired_revision],
													["Applied", collector.applied_revision],
													["Attempted", collector.attempted_revision],
												].map(([label, value]) => (
													<div key={label}>
														<dt className="text-gray-500">{label}</dt>
														<dd className="break-all font-mono">{value || "Unavailable"}</dd>
													</div>
												))}
											</dl>
										</details>
										{collector.config_error &&
											(isPreviousConfigurationError(collector, installation.desired_revision) ? (
												<details className="text-xs text-gray-600">
													<summary className="cursor-pointer">Previous configuration error</summary>
													<p className="mt-2 max-w-[320px] whitespace-pre-wrap break-words rounded bg-gray-50 p-2">
														{collector.config_error}
													</p>
												</details>
											) : (
												<p
													role="alert"
													className="max-w-[320px] whitespace-pre-wrap break-words rounded bg-red-50 p-2 text-xs text-red-800"
												>
													{collector.config_error}
												</p>
											))}
									</td>
									<td className="space-y-2 p-3">
										<dl className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
											<dt className="text-gray-500">CPU</dt>
											<dd>{formatMetric(collector.metrics?.cpu_percent, "%")}</dd>
											<dt className="text-gray-500">Memory</dt>
											<dd>{formatMetric(collector.metrics?.memory_bytes, "MiB", 1024 * 1024)}</dd>
											<dt className="text-gray-500">Logs</dt>
											<dd>{formatMetric(collector.metrics?.logs_sent_per_second, "/ s")}</dd>
											<dt className="text-gray-500">Metric points</dt>
											<dd>{formatMetric(collector.metrics?.metrics_sent_per_second, "/ s")}</dd>
										</dl>
										<p className="text-xs text-gray-500">Sample: {time(collector.metrics_updated_at)}</p>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
		</section>
	);
}

import { useEffect, useState } from "react";
import kubernetesService from "@/services/kubernetes";
import type { KubernetesInstallation } from "@/types/kubernetes.types";

export const refreshKubernetesStatus = () => window.dispatchEvent(new Event("pipelines-updated"));

export function useKubernetesInstallations({
	pipelineId,
	installationId,
	enabled = true,
}: {
	pipelineId?: string;
	installationId?: string;
	enabled?: boolean;
} = {}) {
	const [installations, setInstallations] = useState<KubernetesInstallation[]>([]);
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(enabled);
	const [lastUpdated, setLastUpdated] = useState<number>();
	useEffect(() => {
		if (!enabled) return;
		let active = true;
		let pending = false;
		let controller: AbortController | undefined;
		const refresh = async () => {
			if (pending) return;
			pending = true;
			controller = new AbortController();
			try {
				const data = installationId
					? [await kubernetesService.getInstallation(installationId, controller.signal)]
					: pipelineId
						? await kubernetesService.getPipelineInstallations(pipelineId, controller.signal)
						: await kubernetesService.listInstallations(controller.signal);
				if (active) {
					setInstallations(data ?? []);
					setError("");
					setLastUpdated(Date.now());
				}
			} catch {
				if (active)
					setError(
						"Could not refresh collector status. Retrying every 10 seconds; shown data may be stale.",
					);
			} finally {
				pending = false;
				if (active) setLoading(false);
			}
		};
		setLoading(true);
		void refresh();
		const timer = window.setInterval(() => void refresh(), 10_000);
		window.addEventListener("pipelines-updated", refresh);
		return () => {
			active = false;
			controller?.abort();
			window.clearInterval(timer);
			window.removeEventListener("pipelines-updated", refresh);
		};
	}, [pipelineId, installationId, enabled]);
	return { installations, error, loading, lastUpdated };
}

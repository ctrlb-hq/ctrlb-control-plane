import axiosInstance from "@/utils/axiosInstance";
import { buildManifestRequest } from "@/utils/kubernetes";
import type {
	CreateKubernetesPipeline,
	InstallationCreation,
	KubernetesInstallation,
} from "@/types/kubernetes.types";

async function request<T>(operation: () => Promise<{ data: T }>, fallback: string): Promise<T> {
	try {
		return (await operation()).data;
	} catch (error: unknown) {
		// Axios errors carry request bodies, which can contain credentials.
		// Return only the server's public message, never the Axios error object.
		const data = (error as { response?: { data?: { message?: unknown } } }).response?.data;
		throw new Error(typeof data?.message === "string" ? data.message : fallback);
	}
}

const kubernetesService = {
	createPipeline: (payload: CreateKubernetesPipeline) =>
		request<{ pipeline_id: number }>(
			() => axiosInstance.post("/kubernetes-pipelines", payload, { timeout: 15_000 }),
			"Could not create the Kubernetes pipeline.",
		),
	createInstallation: (payload: { name: string; cluster_name: string; pipeline_id: number }) =>
		request<InstallationCreation>(
			() => axiosInstance.post("/installations", payload, { timeout: 15_000 }),
			"Could not create the installation. Retry this step.",
		),
	listInstallations: (signal?: AbortSignal) =>
		request<KubernetesInstallation[]>(
			() => axiosInstance.get("/installations", { signal, timeout: 8_000 }),
			"Could not refresh Kubernetes installations.",
		),
	getInstallation: (id: string, signal?: AbortSignal) =>
		request<KubernetesInstallation>(
			() => axiosInstance.get(`/installations/${encodeURIComponent(id)}`, { signal, timeout: 8_000 }),
			"Could not refresh the installation.",
		),
	getPipelineInstallations: (id: string, signal?: AbortSignal) =>
		request<KubernetesInstallation[]>(
			() =>
				axiosInstance.get(`/pipelines/${encodeURIComponent(id)}/installations`, {
					signal,
					timeout: 8_000,
				}),
			"Could not refresh collector status.",
		),
	manifest: async (
		id: string,
		token: string,
		backendURL: string,
		image: string,
		signal?: AbortSignal,
	) => {
		const config = buildManifestRequest(id, token, backendURL, image);
		return request<string>(
			() => axiosInstance.request({ ...config, signal, timeout: 15_000 }),
			"Could not generate the manifest. Try again.",
		);
	},
};

export default kubernetesService;

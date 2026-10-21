export interface CollectorMetrics {
	cpu_percent?: number | null;
	memory_bytes?: number | null;
	logs_sent_per_second?: number | null;
	metrics_sent_per_second?: number | null;
	spans_sent_per_second?: number | null;
}

export interface KubernetesCollector {
	id: number;
	installation_id: string;
	node_name: string;
	node_uid: string;
	namespace: string;
	pod_name: string;
	pod_uid: string;
	pod_ip: string;
	version: string;
	platform: string;
	last_seen: number;
	connection_status: "connected" | "disconnected";
	engine_status: "waiting" | "running" | "failed";
	applied_revision: string;
	attempted_revision: string;
	config_error: string;
	rollout_status: "pending" | "applied" | "failed";
	metrics?: CollectorMetrics | null;
	metrics_updated_at?: number | null;
}

export interface RolloutSummary {
	total: number;
	pending: number;
	applied: number;
	failed: number;
	disconnected: number;
}

export interface KubernetesInstallation {
	id: string;
	name: string;
	cluster_name: string;
	pipeline_id: number;
	created_by: string;
	created_at: number;
	desired_revision: string;
	collectors: KubernetesCollector[];
	rollout: RolloutSummary;
}

export interface CreateKubernetesPipeline {
	name: string;
	endpoint: string;
	logs_endpoint?: string;
	metrics_endpoint?: string;
	headers?: Record<string, string>;
	kubelet_insecure_skip_verify?: boolean;
}

export interface InstallationCreation {
	installation: KubernetesInstallation;
	token: string;
}

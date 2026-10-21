import test from "node:test";
import assert from "node:assert/strict";
import {
	buildManifestRequest,
	formatMetric,
	installationCommands,
	isPreviousConfigurationError,
	summarizeInstallations,
} from "./kubernetes.ts";
import { convertToYaml } from "./yaml.ts";
import type { KubernetesCollector, KubernetesInstallation } from "../types/kubernetes.types.ts";

const id = "0123456789abcdef0123456789abcdef";
function collector(overrides: Partial<KubernetesCollector>): KubernetesCollector {
	return {
		id: 1,
		installation_id: id,
		node_name: "worker-1",
		node_uid: "",
		namespace: "observability",
		pod_name: "collector",
		pod_uid: "pod-uid",
		pod_ip: "10.0.0.1",
		version: "0.163.0",
		platform: "linux/amd64",
		last_seen: 1000,
		connection_status: "connected",
		engine_status: "running",
		applied_revision: "old",
		attempted_revision: "desired",
		config_error: "",
		rollout_status: "pending",
		...overrides,
	};
}
function installation(collectors: KubernetesCollector[]): KubernetesInstallation {
	return {
		id,
		name: "demo",
		cluster_name: "demo",
		pipeline_id: 1,
		created_by: "tester",
		created_at: 1000,
		desired_revision: "desired",
		collectors,
		// Summary must come from node records, including failed rollbacks.
		rollout: { total: 999, pending: 999, applied: 999, failed: 999, disconnected: 999 },
	};
}

test("connection and configuration rollout counts remain independent", () => {
	const fleet = [
		installation([
			collector({ id: 1, applied_revision: "desired", rollout_status: "applied" }),
			collector({
				id: 2,
				engine_status: "running",
				config_error: "restored previous config",
				rollout_status: "failed",
			}),
		]),
		installation([
			collector({
				id: 3,
				connection_status: "disconnected",
				engine_status: "waiting",
				rollout_status: "pending",
			}),
			collector({
				id: 4,
				connection_status: "disconnected",
				applied_revision: "desired",
				rollout_status: "applied",
			}),
		]),
	];
	assert.deepEqual(summarizeInstallations(fleet), {
		total: 4,
		connected: 2,
		disconnected: 2,
		pending: 1,
		applied: 2,
		failed: 1,
	});
	assert.deepEqual(summarizeInstallations([installation([])]), {
		total: 0,
		connected: 0,
		disconnected: 0,
		pending: 0,
		applied: 0,
		failed: 0,
	});
});

test("normalized metrics preserve measured zero and distinguish unavailable values", () => {
	assert.equal(formatMetric(0, "/ s"), "0 / s");
	assert.equal(formatMetric(104857600, "MiB", 1024 * 1024), "100 MiB");
	assert.equal(formatMetric(250, "%"), "250 %");
	for (const value of [undefined, null, Number.NaN, Number.POSITIVE_INFINITY, -1])
		assert.equal(formatMetric(value, "%"), "Unavailable");
});

test("only a rejected older revision is historical after successful recovery", () => {
	const recovered = collector({
		applied_revision: "desired",
		attempted_revision: "rejected-candidate",
		config_error: "invalid candidate; retained working configuration",
		rollout_status: "applied",
	});
	assert.equal(isPreviousConfigurationError(recovered, "desired"), true);
	for (const current of [
		{ ...recovered, attempted_revision: "desired", rollout_status: "failed" as const },
		{ ...recovered, attempted_revision: "desired" },
		{ ...recovered, attempted_revision: "" },
		{ ...recovered, applied_revision: "older" },
		{ ...recovered, engine_status: "failed" as const },
		{ ...recovered, rollout_status: "pending" as const },
	]) {
		assert.equal(isPreviousConfigurationError(current, "desired"), false);
	}
});

test("manifest request sends its token only in the POST body", () => {
	const token = "installation-token-with-$-characters";
	const request = buildManifestRequest(
		id,
		token,
		"https://management.example.com/",
		"ctrlb-agent:k8s-v1",
	);
	assert.equal(request.method, "POST");
	assert.equal(request.url, `/installations/${id}/manifest`);
	assert.equal(request.responseType, "text");
	assert.deepEqual(request.data, {
		token,
		backend_url: "https://management.example.com",
		image: "ctrlb-agent:k8s-v1",
	});
	assert.ok(!request.url.includes(token));
	assert.equal(
		buildManifestRequest(id, token, "http://localhost:8096", "local:demo").data.backend_url,
		"http://localhost:8096",
	);
	assert.throws(() => buildManifestRequest(id, token, "", "local:demo"), /Control-plane URL/);
	assert.throws(
		() => buildManifestRequest(id, token, "https://user:password@example.com", "local:demo"),
		/without credentials/,
	);
	assert.throws(
		() => buildManifestRequest(id, token, "https://example.com?token=secret", "local:demo"),
		/without credentials/,
	);
	assert.throws(
		() => buildManifestRequest(id, "", "https://example.com", "local:demo"),
		/token is unavailable/,
	);
	assert.throws(
		() => buildManifestRequest(id, token, "https://example.com", "image with spaces"),
		/image reference/,
	);
});

test("installation commands target one fleet and preserve the shared namespace", () => {
	const commands = installationCommands(id);
	assert.equal(commands.resource, "ctrlb-collector-0123456789ab");
	assert.equal(commands.apply, "kubectl apply -f ctrlb-collector-0123456789ab.yaml");
	assert.ok(
		commands.uninstall.includes("daemonset,secret,serviceaccount ctrlb-collector-0123456789ab"),
	);
	assert.ok(
		commands.uninstall.includes("clusterrole,clusterrolebinding ctrlb-collector-0123456789ab"),
	);
	assert.ok(!commands.uninstall.includes("delete -f"));
	assert.ok(!commands.uninstall.includes("delete namespace"));
	assert.throws(() => installationCommands("unsafe; command"), /Invalid installation/);
});

test("YAML preview retains valid enabled settings", () => {
	const yaml = convertToYaml({
		exporters: {
			otlp_http: { sending_queue: { enabled: false }, retry_on_failure: { enabled: true } },
		},
	});
	assert.match(yaml, /enabled: false/);
	assert.match(yaml, /enabled: true/);
});

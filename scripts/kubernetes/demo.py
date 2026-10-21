#!/usr/bin/env python3
"""Run an isolated Linux kind demo with the control plane on the Docker host."""
import argparse
import copy
import ipaddress
import json
import os
import re
import secrets
import shutil
import signal
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

SCRIPT = Path(__file__).resolve().parent
REPO = SCRIPT.parent.parent


def write_private(path, value):
    text = value if isinstance(value, str) else json.dumps(value, indent=2) + "\n"
    descriptor = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(descriptor, "w") as stream:
        stream.write(text)


def run(*args, capture=False, cwd=None, env=None):
    return subprocess.run(args, check=True, cwd=cwd, env=env, text=True,
                          stdout=subprocess.PIPE if capture else None).stdout


def wait_for(message, predicate, timeout=180):
    print(message, flush=True)
    end = time.monotonic() + timeout
    while time.monotonic() < end:
        result = predicate()
        if result:
            return result
        time.sleep(2)
    raise RuntimeError("Timed out: " + message)


class Demo:
    def __init__(self, directory):
        self.directory = directory.resolve()
        self.path = self.directory / "state.json"
        self.state = json.loads(self.path.read_text()) if self.path.exists() else {}

    def save(self):
        write_private(self.path, self.state)

    def kubectl(self, *args, capture=False):
        return run("kubectl", "--kubeconfig", str(self.directory / "kubeconfig"),
                   "--context", "kind-" + self.state["cluster"], *args, capture=capture)

    def api(self, method, path, body=None, authenticated=True, raw=False, retry=True):
        data = json.dumps(body).encode() if body is not None else None
        headers = {"Content-Type": "application/json"}
        if authenticated:
            headers["Authorization"] = "Bearer " + self.state["access_token"]
        req = urllib.request.Request(self.state["backend_local"] + path, data=data,
                                     headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=15) as response:
                payload = response.read()
                return payload.decode() if raw else json.loads(payload)
        except urllib.error.HTTPError as error:
            # Exception messages deliberately omit headers and request bodies.
            if error.code == 401 and authenticated and retry:
                refreshed = self.api("POST", "/api/auth/v1/refresh", {"refresh_token": self.state["refresh_token"]}, authenticated=False)
                self.state["access_token"] = refreshed["access_token"]
                self.save()
                return self.api(method, path, body, authenticated, raw, retry=False)
            raise RuntimeError(f"API {method} {path} returned HTTP {error.code}") from None

    def installation(self):
        return self.api("GET", "/api/frontend/v2/installations/" + self.state["installation_id"])

    def start_process(self, name, command, env=None):
        with (self.directory / (name + ".log")).open("ab") as log:
            process = subprocess.Popen(command, cwd=self.directory, env=env,
                                       stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        self.state[name + "_pid"] = process.pid
        self.state[name + "_command"] = command
        self.save()

    def stop_process(self, name):
        pid = self.state.get(name + "_pid")
        if not pid:
            return
        command = self.state.get(name + "_command", [])
        proc = Path(f"/proc/{pid}/cmdline")
        if proc.exists():
            actual = proc.read_bytes().split(b"\0")
            expected = [str(part).encode() for part in command]
            if actual[:len(expected)] != expected:
                raise RuntimeError(f"Refusing to stop reused PID for {name}")
            os.kill(pid, signal.SIGTERM)
            end = time.monotonic() + 15
            while proc.exists() and time.monotonic() < end:
                status = Path(f"/proc/{pid}/stat").read_text().split()[2]
                if status == "Z":
                    break
                time.sleep(0.2)
        self.state.pop(name + "_pid", None)
        self.save()

    def start_backend(self):
        env = dict(os.environ, JWT_SECRET=self.state["jwt_secret"], PORT=str(self.state["backend_port"]), ENV="dev")
        self.start_process("backend", [str(self.directory / "backend")], env)
        def listening():
            try:
                with socket.create_connection(("127.0.0.1", self.state["backend_port"]), timeout=1):
                    return True
            except OSError:
                return False
        wait_for("Waiting for the host control plane", listening, timeout=30)

    def wait_applied(self, different_from=None):
        def applied():
            view = self.installation()
            return view if (view["rollout"]["total"] == 2 and view["rollout"]["applied"] == 2 and view["rollout"]["disconnected"] == 0
                            and (different_from is None or view["desired_revision"] != different_from)) else None
        return wait_for("Waiting for both node collectors to apply the desired revision", applied)

    def up(self, args):
        if not (1 <= args.backend_port <= 65535 and 1 <= args.receiver_port <= 65535) or args.backend_port == args.receiver_port:
            raise RuntimeError("Demo ports must be different valid TCP ports")
        auth_value = None
        if args.real:
            auth_path = os.environ.get("CTRLB_AUTH_FILE")
            if not auth_path:
                raise RuntimeError("--real requires CTRLB_AUTH_FILE containing the literal Authorization header value")
            auth_value = Path(auth_path).read_text().strip()
            if not auth_value or "\n" in auth_value or "\r" in auth_value:
                raise RuntimeError("CTRLB_AUTH_FILE must contain one nonempty header value")
        for tool in ("docker", "kind", "kubectl", "go"):
            if not shutil.which(tool) and not (tool == "go" and os.environ.get("BACKEND_BINARY") and os.environ.get("COLLECTOR_BINARY")):
                raise RuntimeError("Missing required tool: " + tool)
        if self.state:
            raise RuntimeError("This demo directory already has state; use status, verify, or down")
        cluster = os.environ.get("DEMO_CLUSTER", "ctrlb-v1-demo")
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,40}", cluster):
            raise RuntimeError("DEMO_CLUSTER must be a short lowercase Kubernetes name")
        if cluster in run("kind", "get", "clusters", capture=True).splitlines():
            raise RuntimeError("Refusing to reuse an existing kind cluster: " + cluster)
        self.directory.mkdir(mode=0o700, parents=True, exist_ok=False)
        self.state = {"cluster": cluster, "backend_port": args.backend_port, "receiver_port": args.receiver_port,
                      "backend_local": f"http://127.0.0.1:{args.backend_port}", "jwt_secret": secrets.token_urlsafe(32),
                      "image": os.environ.get("COLLECTOR_IMAGE", "ctrlb-agent:k8s-v1"), "real_destination": args.real}
        self.save()
        source_backend = os.environ.get("BACKEND_BINARY")
        if source_backend:
            shutil.copyfile(source_backend, self.directory / "backend")
        else:
            run("go", "build", "-o", str(self.directory / "backend"), "./cmd/backend", cwd=REPO / "backend")
        (self.directory / "backend").chmod(0o755)
        run("bash", str(SCRIPT / "build-image.sh"))
        run("kind", "create", "cluster", "--name", cluster, "--config", str(SCRIPT / "kind.yaml"),
            "--kubeconfig", str(self.directory / "kubeconfig"), "--wait", "180s")
        run("kind", "load", "docker-image", "--name", cluster, self.state["image"])
        network = json.loads(run("docker", "network", "inspect", "kind", capture=True))[0]
        gateways = [entry.get("Gateway", "") for entry in network["IPAM"]["Config"]]
        gateway = next((value for value in gateways if value and ipaddress.ip_address(value).version == 4), None)
        if gateway is None:
            raise RuntimeError("The Linux demo requires an IPv4 gateway on the kind Docker network")
        self.state["backend_cluster"] = f"http://{gateway}:{args.backend_port}"
        self.save()
        self.start_process("receiver", [sys.executable, str(SCRIPT / "otlp_receiver.py"), "--port", str(args.receiver_port),
                                        "--evidence", str(self.directory / "otlp-evidence.json")])
        self.start_backend()
        account = {"name": "Kubernetes demo", "email": f"demo-{secrets.token_hex(4)}@example.com", "password": secrets.token_urlsafe(20)}
        auth = self.api("POST", "/api/auth/v1/register", account, authenticated=False)
        self.state.update(access_token=auth["access_token"], refresh_token=auth["refresh_token"])
        write_private(self.directory / "login.json", account)
        self.save()
        pipeline = {"name": "Kubernetes v1 demo", "endpoint": f"http://{gateway}:{args.receiver_port}", "kubelet_insecure_skip_verify": True}
        if args.real:
            pipeline.update(endpoint=os.environ.get("CTRLB_OTLP_ENDPOINT", "https://ctrlb.explore.ctrlb.ai/engine/api/default"),
                            logs_endpoint=os.environ.get("CTRLB_LOGS_ENDPOINT", "https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs"),
                            metrics_endpoint=os.environ.get("CTRLB_METRICS_ENDPOINT", "https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics"),
                            headers={"Authorization": auth_value})
        result = self.api("POST", "/api/frontend/v2/kubernetes-pipelines", pipeline)
        self.state["pipeline_id"] = result["pipeline_id"]
        created = self.api("POST", "/api/frontend/v2/installations", {"name": "Kubernetes v1 demo", "cluster_name": cluster, "pipeline_id": result["pipeline_id"]})
        self.state["installation_id"] = created["installation"]["id"]
        self.state["installation_token"] = created["token"]
        self.state["resource_name"] = "ctrlb-collector-" + created["installation"]["id"][:12]
        self.save()
        manifest = self.api("POST", "/api/frontend/v2/installations/" + self.state["installation_id"] + "/manifest",
                            {"token": created["token"], "backend_url": self.state["backend_cluster"], "image": self.state["image"]}, raw=True)
        write_private(self.directory / "collector.yaml", manifest)
        self.kubectl("apply", "-f", str(self.directory / "collector.yaml"))
        self.kubectl("apply", "-f", str(SCRIPT / "workload.yaml"))
        self.kubectl("rollout", "status", "daemonset/" + self.state["resource_name"], "-n", "observability", "--timeout=180s")
        self.kubectl("rollout", "status", "daemonset/log-generator", "-n", "ctrlb-demo", "--timeout=180s")
        self.wait_applied()
        print("Demo ready. Private state and logs: " + str(self.directory))
        print("Run demo.sh --dir " + str(self.directory) + " verify")

    def evidence(self):
        path = self.directory / "otlp-evidence.json"
        return json.loads(path.read_text()) if path.exists() else {}

    def verify(self):
        initial = self.wait_applied()
        ids = {c["node_name"]: c["id"] for c in initial["collectors"]}
        if not self.state.get("real_destination"):
            def telemetry():
                evidence = self.evidence()
                nodes = {attrs.get("k8s.node.name") for attrs in evidence.get("resources", [])}
                metric_names = evidence.get("metric_names", [])
                return (set(ids) <= nodes and any("ctrlb-demo-marker" in body for body in evidence.get("log_bodies", []))
                        and all(any(name.startswith(prefix) for name in metric_names) for prefix in ("k8s.node.", "k8s.pod.", "container.")))
            wait_for("Checking received pod logs, both node identities, and node/pod/container metrics", telemetry)
        graph_path = f"/api/frontend/v2/pipelines/{self.state['pipeline_id']}/graph"
        good_graph = self.api("GET", graph_path)
        for node in good_graph["nodes"]:
            if node["component_name"] == "resource_processor":
                attrs = node["config"].setdefault("attributes", [])
                attrs[:] = [a for a in attrs if a.get("key") != "ctrlb.demo.revision"]
                attrs.append({"key": "ctrlb.demo.revision", "value": str(time.time_ns()), "action": "upsert"})
        self.api("POST", graph_path, good_graph)
        updated = self.wait_applied(different_from=initial["desired_revision"])
        if updated["desired_revision"] == initial["desired_revision"]:
            raise RuntimeError("Configuration edit did not produce a new revision")
        good_revision = updated["desired_revision"]
        bad_graph = copy.deepcopy(good_graph)
        invalidated_receiver = False
        for node in bad_graph["nodes"]:
            if node["component_role"] == "receiver" and node["component_name"] == "kubeletstats_receiver":
                # Use an invalid typed value: unknown keys may be ignored by the
                # Collector's config decoder and do not prove rejection.
                node["config"]["collection_interval"] = "invalid-duration-demo"
                invalidated_receiver = True
        if not invalidated_receiver:
            raise RuntimeError("Cannot verify rollback: pipeline has no kubeletstats receiver")
        self.api("POST", graph_path, bad_graph)
        def rolled_back():
            view = self.installation()
            return view if view["rollout"]["failed"] == 2 else None
        failed = wait_for("Checking rejected configuration and rollback on both nodes", rolled_back)
        for collector in failed["collectors"]:
            if collector["engine_status"] != "running" or collector["applied_revision"] != good_revision or not collector["config_error"]:
                raise RuntimeError("A failed candidate did not retain the running configuration")
        self.api("POST", graph_path, good_graph)
        self.wait_applied()
        self.replace_one(ids, outage=False)
        self.replace_one(ids, outage=True)
        report = {"checks": ["two collectors share one pipeline", "config update applied to both nodes", "invalid config rejected with working config retained",
                             "pod replacement preserves collector IDs", "cached config starts while control plane is offline", "reconnection preserves collector IDs"],
                  "installation": self.installation(), "local_otlp": None if self.state.get("real_destination") else self.evidence()}
        write_private(self.directory / "verification.json", report)
        print("Lifecycle verification passed. Evidence: " + str(self.directory / "verification.json"))
        if self.state.get("real_destination"):
            print("Confirm the identifiable pod logs and Kubernetes metric series in CtrlB; HTTP acceptance alone does not prove indexing.")

    def replace_one(self, ids, outage):
        previous = self.installation()
        collector = previous["collectors"][0]
        if outage:
            self.stop_process("backend")
        try:
            self.kubectl("delete", "pod", collector["pod_name"], "-n", "observability", "--wait=true")
            self.kubectl("rollout", "status", "daemonset/" + self.state["resource_name"], "-n", "observability", "--timeout=180s")
            if outage:
                # Ready without the backend requires restoring a cached running config.
                print("Replacement collector became ready while the control plane was stopped", flush=True)
        finally:
            if outage:
                self.start_backend()
        def replaced():
            view = self.installation()
            current = {c["node_name"]: c["id"] for c in view["collectors"]}
            if current != ids:
                raise RuntimeError("Pod replacement changed stable collector identity")
            return view if (view["rollout"]["applied"] == 2 and view["rollout"]["disconnected"] == 0
                            and any(c["node_name"] == collector["node_name"] and c["pod_uid"] != collector["pod_uid"] for c in view["collectors"])) else None
        wait_for("Checking stable identity after pod replacement" + (" and control-plane outage" if outage else ""), replaced)

    def down(self, delete_cluster):
        name = self.state.get("resource_name")
        if name:
            # Never delete the downloaded manifest: it includes a shared namespace.
            self.kubectl("delete", "daemonset,secret,serviceaccount", name, "-n", "observability", "--ignore-not-found")
            self.kubectl("delete", "clusterrole,clusterrolebinding", name, "--ignore-not-found")
            self.kubectl("get", "namespace", "observability")
            self.kubectl("delete", "daemonset/log-generator", "-n", "ctrlb-demo", "--ignore-not-found")
            print("Installation removed; observability namespace and other installations preserved")
        self.stop_process("backend")
        self.stop_process("receiver")
        if delete_cluster:
            run("kind", "delete", "cluster", "--name", self.state["cluster"], "--kubeconfig", str(self.directory / "kubeconfig"))
        print("Private demo files retained at " + str(self.directory))


def main():
    os.umask(0o077)
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dir", type=Path, default=Path(os.environ.get("DEMO_DIR", "/tmp/ctrlb-k8s-demo")))
    commands = parser.add_subparsers(dest="command", required=True)
    up = commands.add_parser("up")
    up.add_argument("--backend-port", type=int, default=18096)
    up.add_argument("--receiver-port", type=int, default=14318)
    up.add_argument("--real", action="store_true", help="export to CtrlB using CTRLB_AUTH_FILE")
    commands.add_parser("verify")
    commands.add_parser("status")
    down = commands.add_parser("down")
    down.add_argument("--delete-cluster", action="store_true")
    args = parser.parse_args()
    demo = Demo(args.dir)
    if args.command != "up" and not demo.state:
        parser.error("No demo state in the selected directory")
    if args.command == "up":
        demo.up(args)
    elif args.command == "verify":
        demo.verify()
    elif args.command == "status":
        print(json.dumps(demo.installation(), indent=2))
    else:
        demo.down(args.delete_cluster)


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, OSError, subprocess.CalledProcessError) as error:
        print("Demo failed: " + str(error), file=sys.stderr)
        sys.exit(1)

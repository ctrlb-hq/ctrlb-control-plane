import json
import socket
import subprocess
import sys
import tempfile
import time
import unittest
import urllib.request
from pathlib import Path

from otlp_receiver import fields, summarize


def varint(value):
    data = bytearray()
    while value > 127:
        data.append((value & 127) | 128)
        value >>= 7
    return bytes(data + bytes([value]))


def message(number, value):
    return varint(number << 3 | 2) + varint(len(value)) + value


def attr(key, value):
    return message(1, message(1, key.encode()) + message(2, message(1, value.encode())))


def request(signal):
    resource = attr("k8s.node.name", "worker-1") + attr("k8s.namespace.name", "ctrlb-demo")
    if signal == "logs":
        item = message(5, message(1, b"ctrlb-demo-marker"))
    else:
        item = message(1, b"k8s.pod.cpu.usage")
    return message(1, message(1, resource) + message(2, message(2, item)))


class OTLPReceiverTests(unittest.TestCase):
    def test_logs_and_metrics_keep_resource_metadata(self):
        resources, logs, metrics = summarize(request("logs"), "logs")
        self.assertEqual(resources, [{"k8s.node.name": "worker-1", "k8s.namespace.name": "ctrlb-demo"}])
        self.assertEqual(logs, ["ctrlb-demo-marker"])
        self.assertFalse(metrics)
        resources, logs, metrics = summarize(request("metrics"), "metrics")
        self.assertFalse(logs)
        self.assertEqual(metrics, {"k8s.pod.cpu.usage"})

    def test_truncated_and_invalid_protobuf_is_rejected(self):
        for data in (b"\x0a\x80", b"\x0a\x03a", b"\x00", b"\x0b"):
            with self.assertRaises(ValueError):
                fields(data)

    def test_http_receiver_records_evidence_without_authorization(self):
        with tempfile.TemporaryDirectory() as temporary:
            evidence = Path(temporary) / "evidence.json"
            with socket.socket() as sock:
                sock.bind(("127.0.0.1", 0))
                port = sock.getsockname()[1]
            process = subprocess.Popen([sys.executable, str(Path(__file__).with_name("otlp_receiver.py")), "--port", str(port), "--evidence", str(evidence)],
                                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            try:
                deadline = time.monotonic() + 5
                while time.monotonic() < deadline:
                    if process.poll() is not None:
                        self.fail("OTLP receiver exited before its listener became ready")
                    try:
                        with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                            break
                    except OSError:
                        time.sleep(0.01)
                else:
                    self.fail("OTLP receiver listener was not ready within five seconds")
                req = urllib.request.Request(f"http://127.0.0.1:{port}/v1/logs", data=request("logs"),
                                             headers={"Content-Type": "application/x-protobuf", "Authorization": "Basic never-store-this"})
                with urllib.request.urlopen(req) as response:
                    self.assertEqual(response.status, 200)
                    self.assertEqual(response.read(), b"")
                stored = json.loads(evidence.read_text())
                self.assertEqual(stored["requests"]["logs"], 1)
                self.assertEqual(stored["log_bodies"], ["ctrlb-demo-marker"])
                self.assertNotIn("never-store-this", evidence.read_text())
                self.assertEqual(evidence.stat().st_mode & 0o777, 0o600)
            finally:
                process.terminate()
                stdout, stderr = process.communicate(timeout=5)
                self.assertNotIn(b"never-store-this", stdout + stderr)


if __name__ == "__main__":
    unittest.main()

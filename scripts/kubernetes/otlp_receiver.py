#!/usr/bin/env python3
"""Small local OTLP/protobuf receiver for demo evidence, not a telemetry store."""
import argparse
import gzip
import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path


def fields(data):
    """Decode protobuf wire fields without a generated-code dependency."""
    def varint(offset):
        value = 0
        for shift in range(0, 70, 7):
            if offset >= len(data):
                raise ValueError("truncated protobuf")
            byte = data[offset]
            offset += 1
            value |= (byte & 127) << shift
            if byte < 128:
                return value, offset
        raise ValueError("invalid varint")

    result, offset = {}, 0
    while offset < len(data):
        tag, offset = varint(offset)
        number, wire = tag >> 3, tag & 7
        if number == 0:
            raise ValueError("invalid protobuf field")
        if wire == 0:
            value, offset = varint(offset)
        elif wire in (1, 5):
            size = 8 if wire == 1 else 4
            value, offset = data[offset:offset + size], offset + size
            if len(value) != size:
                raise ValueError("truncated fixed-width value")
        elif wire == 2:
            size, offset = varint(offset)
            value, offset = data[offset:offset + size], offset + size
            if len(value) != size:
                raise ValueError("truncated length-delimited value")
        else:
            raise ValueError("unsupported protobuf wire type")
        result.setdefault(number, []).append(value)
    return result


def text_value(data):
    values = fields(data)
    if 1 in values:
        return values[1][0].decode("utf-8", errors="replace")
    if 2 in values:
        return bool(values[2][0])
    if 3 in values:
        return values[3][0]
    return None


def resource_attributes(data):
    attrs = {}
    for item in fields(data).get(1, []):
        kv = fields(item)
        key = kv.get(1, [b""])[0].decode("utf-8", errors="replace")
        if 2 in kv:
            attrs[key] = text_value(kv[2][0])
    return attrs


def summarize(data, signal):
    resources, logs, metrics = [], [], set()
    for batch in fields(data).get(1, []):
        resource_batch = fields(batch)
        resources.append(resource_attributes(resource_batch.get(1, [b""])[0]))
        for scope in resource_batch.get(2, []):
            for item in fields(scope).get(2, []):
                item_fields = fields(item)
                if signal == "logs":
                    logs.append(text_value(item_fields[5][0]) if 5 in item_fields else None)
                elif 1 in item_fields:
                    metrics.add(item_fields[1][0].decode("utf-8", errors="replace"))
    return resources, logs, metrics


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--port", type=int, default=14318)
    parser.add_argument("--evidence", type=Path, required=True)
    args = parser.parse_args()
    lock = threading.Lock()
    evidence = {"requests": {"logs": 0, "metrics": 0}, "resources": [], "log_bodies": [], "metric_names": []}

    def save():
        temporary = args.evidence.with_suffix(".tmp")
        temporary.write_text(json.dumps(evidence, indent=2) + "\n")
        temporary.chmod(0o600)
        temporary.replace(args.evidence)

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass  # Requests may carry credentials. Never log headers or payloads.

        def do_GET(self):
            if self.path != "/evidence":
                self.send_error(404)
                return
            with lock:
                body = json.dumps(evidence).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def do_POST(self):
            signal = self.path.rsplit("/", 1)[-1]
            if signal not in ("logs", "metrics"):
                self.send_error(404)
                return
            try:
                size = int(self.headers.get("Content-Length", "0"))
                if size < 0 or size > 10 * 1024 * 1024:
                    self.send_error(413)
                    return
                data = self.rfile.read(size)
                if self.headers.get("Content-Encoding") == "gzip":
                    data = gzip.decompress(data)
                    if len(data) > 10 * 1024 * 1024:
                        raise ValueError("expanded request too large")
                resources, logs, metrics = summarize(data, signal)
                with lock:
                    evidence["requests"][signal] += 1
                    # Bound this receiver's memory and retain inspectable metadata.
                    for attrs in resources:
                        if attrs not in evidence["resources"] and len(evidence["resources"]) < 1000:
                            evidence["resources"].append(attrs)
                    evidence["log_bodies"] = (evidence["log_bodies"] + [v for v in logs if isinstance(v, str)])[-200:]
                    evidence["metric_names"] = sorted(set(evidence["metric_names"]) | metrics)
                    save()
            except (ValueError, OSError, EOFError):
                self.send_error(400, "Invalid OTLP protobuf request")
                return
            self.send_response(200)
            self.send_header("Content-Type", "application/x-protobuf")
            self.send_header("Content-Length", "0")
            self.end_headers()

    args.evidence.parent.mkdir(parents=True, exist_ok=True)
    with lock:
        save()
    ThreadingHTTPServer(("0.0.0.0", args.port), Handler).serve_forever()


if __name__ == "__main__":
    os.umask(0o077)
    main()

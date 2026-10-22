import test from "node:test";
import assert from "node:assert/strict";
import { resolveApiBaseUrl } from "./urls.ts";

const hosted = {
	appBasePath: "/control-plane/",
	origin: "https://ctrlb.explore.ctrlb.ai",
};

test("hosted API URLs preserve the application prefix and current origin", () => {
	assert.equal(resolveApiBaseUrl(hosted), "https://ctrlb.explore.ctrlb.ai/control-plane/api");
	assert.equal(
		resolveApiBaseUrl({ ...hosted, origin: "https://staging.example.com:8443" }),
		"https://staging.example.com:8443/control-plane/api",
	);
	assert.equal(
		resolveApiBaseUrl({ ...hosted, appBasePath: "/" }),
		"https://ctrlb.explore.ctrlb.ai/api",
	);
});

test("explicit API roots work with a relative prefix or a separate backend origin", () => {
	assert.equal(
		resolveApiBaseUrl({ ...hosted, apiBaseUrl: "/control-plane/api/" }),
		"https://ctrlb.explore.ctrlb.ai/control-plane/api",
	);
	assert.equal(
		resolveApiBaseUrl({
			...hosted,
			apiBaseUrl: "https://backend.example.com/custom/api/",
			legacyBackendUrl: "http://localhost:8096",
		}),
		"https://backend.example.com/custom/api",
	);
});

test("legacy backend roots and local development keep their existing API routes", () => {
	assert.equal(
		resolveApiBaseUrl({ ...hosted, legacyBackendUrl: "http://localhost:18096/" }),
		"http://localhost:18096/api",
	);
	assert.equal(
		resolveApiBaseUrl({ ...hosted, legacyBackendUrl: "https://backend.example.com/control-plane" }),
		"https://backend.example.com/control-plane/api",
	);
	assert.equal(
		resolveApiBaseUrl({ ...hosted, apiBaseUrl: " ", legacyBackendUrl: "", development: true }),
		"http://localhost:8096/api",
	);
});

test("API roots reject credentials, query strings, fragments and non-HTTP URLs", () => {
	for (const apiBaseUrl of [
		"https://user:password@example.com/api",
		"https://example.com/api?token=secret",
		"https://example.com/api#fragment",
		"file:///api",
	]) {
		assert.throws(() => resolveApiBaseUrl({ ...hosted, apiBaseUrl }), /API URL must use HTTP/);
	}
});

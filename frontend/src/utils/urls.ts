interface ApiUrlOptions {
	apiBaseUrl?: string;
	legacyBackendUrl?: string;
	appBasePath: string;
	origin: string;
	development?: boolean;
}

export function resolveApiBaseUrl({
	apiBaseUrl,
	legacyBackendUrl,
	appBasePath,
	origin,
	development = false,
}: ApiUrlOptions): string {
	const explicit = apiBaseUrl?.trim();
	const legacy = legacyBackendUrl?.trim();
	const configured = explicit || (legacy ? `${legacy.replace(/\/+$/, "")}/api` : "");
	const fallback = development
		? "http://localhost:8096/api"
		: `${appBasePath.replace(/\/+$/, "")}/api`;
	const url = new URL(configured || fallback, origin);
	if (
		!["http:", "https:"].includes(url.protocol) ||
		url.username ||
		url.password ||
		url.search ||
		url.hash
	) {
		throw new Error("API URL must use HTTP or HTTPS without credentials, a query, or a fragment.");
	}
	return url.href.replace(/\/+$/, "");
}

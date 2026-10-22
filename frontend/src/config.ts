import { resolveApiBaseUrl } from "./utils/urls";

// VITE_API_BASE_URL includes /api. VITE_API_URL remains the legacy backend root.
export const API_BASE_URL = resolveApiBaseUrl({
	apiBaseUrl: import.meta.env.VITE_API_BASE_URL,
	legacyBackendUrl: import.meta.env.VITE_API_URL,
	appBasePath: import.meta.env.BASE_URL,
	origin: window.location.origin,
	development: import.meta.env.DEV,
});

export const apiUrl = (path: string): string => `${API_BASE_URL}/${path.replace(/^\/+/, "")}`;
export const FRONTEND_API_URL = apiUrl("frontend/v2");
export const AUTH_API_URL = apiUrl("auth/v1");

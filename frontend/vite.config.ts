import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
	base: loadEnv(mode, process.cwd(), "VITE_").VITE_BASE_PATH || "/",
	plugins: [react()],
	server: {
		port: 3030,
	},
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "./src"),
		},
	},
}));

import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), "");
	const clientPort = Number(env.VITE_DEV_PORT || "3001");
	const apiProxyTarget = env.VITE_API_PROXY_TARGET || "http://127.0.0.1:8080";

	return {
		plugins: [tailwindcss(), reactRouter()],
		resolve: { tsconfigPaths: true },
		server: {
			port: clientPort,
			proxy: {
				"/api": { target: apiProxyTarget, changeOrigin: true },
				"/ws": { target: apiProxyTarget, ws: true },
			},
		},
	};
});

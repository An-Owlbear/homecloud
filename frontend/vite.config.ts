import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import basicSsl from '@vitejs/plugin-basic-ssl';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss(), basicSsl()],
	server: {
		host: 'hc.anowlbear.com',
		proxy: {
			'/api': 'http://hc.anowlbear.com:1323',
			'/assets/data': 'http://hc.anowlbear.com:1323'
		}
	}
});

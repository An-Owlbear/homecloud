import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	server: {
		host: 'hc.anowlbear.com',
		proxy: {
			'/api': 'http://hc.anowlbear.com:1323',
			'/assets/data': 'http://hc.anowlbear.com:1323'
		}
	}
});

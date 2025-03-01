import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: 'hc.anowlbear.com',
		proxy: {
			'/api': 'http://hc.anowlbear.com:1323'
		}
	}
});

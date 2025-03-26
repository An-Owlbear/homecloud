import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import basicSsl from '@vitejs/plugin-basic-ssl';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss(), basicSsl()],
	server: {
		host: 'myserver.homecloudapp.com',
		proxy: {
			'/api': 'https://myserver.homecloudapp.com',
			'/assets/data': 'https://myserver.homecloudapp.com',
		}
	}
});

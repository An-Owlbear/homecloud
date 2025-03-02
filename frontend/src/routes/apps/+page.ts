import type { PageLoad } from './$types';
import type { HomecloudApp } from '$lib/models';
import { goto } from '$app/navigation';

export const load: PageLoad = async ({ fetch }) => {
	const response = await fetch('/api/v1/apps');
	console.log(response);
	if (!response.ok) {
		await goto('/auth/login');
	}

	const apps = await response.json() as HomecloudApp[];
	return {
		apps: apps,
	};
}
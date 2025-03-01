import type { HomecloudApp } from '$lib/models';

export const GetApps = async (): Promise<HomecloudApp[]> => {
	const response = await fetch('/api/v1/apps');
	if (!response.ok) {
		throw new Error(response.statusText);
	}

	return await response.json() as HomecloudApp[];
}
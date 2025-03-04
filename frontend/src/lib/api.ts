import type { HomecloudApp, PackageListItem } from '$lib/models';
import { goto } from '$app/navigation';

export const GetApps = async (): Promise<HomecloudApp[]> => {
	const response = await fetch('/api/v1/apps');
	if (!response.ok) {
		throw new Error(response.statusText);
	}

	return await response.json() as HomecloudApp[];
}

export const searchPackages = async (search: string): Promise<PackageListItem[]> => {
	const response = await fetch('/api/v1/packages/search?' + new URLSearchParams({ q: search}));
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}

	return await response.json() as PackageListItem[];
}

export const CheckAuthRedirect = async (response: Response) => {
	if (response.status === 401) {
		await goto('/auth/login');
	}
}
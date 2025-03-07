import type { HomecloudApp, PackageListItem, SearchParams, User } from '$lib/models';
import { goto } from '$app/navigation';

export const GetApps = async (): Promise<HomecloudApp[]> => {
	const response = await fetch('/api/v1/apps');
	if (!response.ok) {
		throw new Error(response.statusText);
	}

	return await response.json() as HomecloudApp[];
}

export const searchPackages = async (params: SearchParams): Promise<PackageListItem[]> => {
	const response = await fetch('/api/v1/packages/search?' + new URLSearchParams(params));
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}

	return await response.json() as PackageListItem[];
}

export const getPackage = async (id: string): Promise<PackageListItem> => {
	const response = await fetch(`/api/v1/packages/${id}`);
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}

	return await response.json() as PackageListItem;
}

export const getUsers = async (): Promise<User[]> => {
	const response = await fetch('/api/v1/users');
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
	return await response.json() as User[];
}

export const CheckAuthRedirect = async (response: Response) => {
	if (response.status === 401) {
		await goto('/auth/login');
	}
}
import type { HomecloudApp, InviteCode, PackageListItem, SearchParams, User } from '$lib/models';
import { goto } from '$app/navigation';

export const getApps = async (): Promise<HomecloudApp[]> => {
	const response = await fetch('/api/v1/apps');
	if (!response.ok) {
		await CheckAuthRedirect(response);
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

export const installPackage = async (id: string): Promise<void> => {
	const response = await fetch(`/api/v1/packages/${id}/install`, { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
}

export const uninstallApp = async (id: string): Promise<void> => {
	const response = await fetch(`/api/v1/apps/${id}/uninstall`, { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
}

export const getUsers = async (): Promise<User[]> => {
	const response = await fetch('/api/v1/users');
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
	return await response.json() as User[];
}

export const deleteUser = async (id: string): Promise<void> => {
	const response = await fetch(`/api/v1/users/${id}`, { method: 'DELETE' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
}

export const inviteUser = async (): Promise<InviteCode> => {
	const response = await fetch('/api/v1/invites', { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
	return await response.json() as InviteCode;
}

export const CheckAuthRedirect = async (response: Response) => {
	if (response.status === 401) {
		await goto('/auth/login');
	}
}
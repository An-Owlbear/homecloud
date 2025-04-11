import type {
	ExternalStorage,
	HomecloudApp,
	InviteCode,
	PackageListItem, RecoveryCode,
	SearchParams, StoreHome,
	UpdateCheckResponse, UpdateUserOptions,
	User, UserOptions
} from '$lib/models';
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

export const checkUpdates = async (): Promise<UpdateCheckResponse> => {
	const response = await fetch('/api/v1/update');
	if (!response.ok) {
		await CheckAuthRedirect(response);
		// throw new Error(response.statusText);
	}
	return await response.json() as UpdateCheckResponse;
}

export const updateSystem = async (): Promise<void> => {
	const response = await fetch('/api/v1/update', { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}
}

export const listExternalStorage = async (): Promise<ExternalStorage[]> => {
	const response = await fetch('/api/v1/backup/devices');
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
	return await response.json() as ExternalStorage[];
}

export const listBackups = async (externalStorage: string, appId: string): Promise<string[]> => {
	const response = await fetch(`/api/v1/apps/${appId}/backups?` + new URLSearchParams({ target_device: externalStorage }));
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}

	return await response.json() as string[];
}

export const backupApp = async (externalStorage: string, appId: string): Promise<void> => {
	const response = await fetch(`/api/v1/apps/${appId}/backup`, {
		method: 'POST',
		body: JSON.stringify({ target_device: externalStorage }),
		headers: { 'Content-Type': 'application/json' },
	});
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
}

export const restoreBackup = async (externalStorage: string, appId: string, backup: string): Promise<void> => {
	const response = await fetch(`/api/v1/apps/${appId}/restore`, {
		method: 'POST',
		body: JSON.stringify({ target_device: externalStorage, backup: backup }),
		headers: { 'Content-Type': 'application/json' },
	});
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
}

export const getUserOptions = async (): Promise<UserOptions> => {
	const response = await fetch('/api/v1/account/options');
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
	return await response.json() as UserOptions;
}

export const updateUserOptions = async (options: UpdateUserOptions): Promise<void> => {
	const response = await fetch('/api/v1/account/options', {
		method: 'PUT',
		body: JSON.stringify(options),
		headers: { 'Content-Type': 'application/json' },
	});
}

export const getStoreHome = async () => {
	const response = await fetch('/api/v1/store');
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
	return await response.json() as StoreHome;
}

export const createRecoveryCode  = async (userId: string): Promise<RecoveryCode> => {
	const response = await fetch(`/api/v1/users/${userId}/reset_password`, { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
	return await response.json() as RecoveryCode;
}

export const updatePackageList = async () => {
	const response = await fetch('/api/v1/packages/update', { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
}

export const getAvailableUpdates = async (): Promise<PackageListItem[]> => {
	const response = await fetch('/api/v1/apps/update');
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
	return await response.json() as Promise<PackageListItem[]>;
}

export const updateAppsRequest = async () => {
	const response = await fetch('/api/v1/apps/update', { method: 'POST' });
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}
}

export const CheckAuthRedirect = async (response: Response) => {
	if (response.status === 401) {
		await goto('/auth/login');
	}
}
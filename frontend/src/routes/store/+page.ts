import type { PageLoad } from './$types';
import { CheckAuthRedirect } from '$lib/api';
import type { PackageListItem } from '$lib/models';

export const load: PageLoad = async ({ fetch }) => {
	const response = await fetch('/api/v1/packages');
	if (!response.ok) {
		await CheckAuthRedirect(response);
	}

	const packages = await response.json() as PackageListItem[];
	return {
		packages: packages
	}
}
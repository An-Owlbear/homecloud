import type { PageLoad } from './$types'
import { CheckAuthRedirect } from '$lib/api';
import type { PackageListItem } from '$lib/models';

export const load: PageLoad = async ({ fetch, params }) => {
	const response = await fetch('/api/v1/packages/search?' + new URLSearchParams({ category: params.category }));
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}

	return {
		packages: await response.json() as PackageListItem[]
	}
}
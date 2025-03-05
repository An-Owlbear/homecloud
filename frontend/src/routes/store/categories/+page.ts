import type { PageLoad } from './$types';
import { CheckAuthRedirect } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const response = await fetch('/api/v1/packages/categories');
	if (!response.ok) {
		await CheckAuthRedirect(response);
		throw new Error(response.statusText);
	}

	return {
		categories: await response.json() as string[]
	};
}
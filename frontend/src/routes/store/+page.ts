import type { PageLoad } from './$types';
import { searchPackages } from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const packages = await searchPackages(url.searchParams.get('q') ?? '');
	return {
		packages: packages
	}
}
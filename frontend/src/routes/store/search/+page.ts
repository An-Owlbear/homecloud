import type { PageLoad } from './$types';
import { searchPackages } from '$lib/api';

export const load: PageLoad = async ({ url }) => {
	const searchTerm = url.searchParams.get('q') ?? '';
	const packages = await searchPackages({ q: searchTerm });
	return {
		packages: packages
	}
}
import type { PageLoad } from './$types';
import { searchPackages } from '$lib/api';

export const load: PageLoad = async ({ params }) => {
	const packages = await searchPackages({ developer: params.id });
	return {
		packages: packages
	}
}
import type { PageLoad } from './$types';
import { getPackage } from '$lib/api';

export const load: PageLoad = async ({ params }) => {
	return {
		package: await getPackage(params.id)
	}
}
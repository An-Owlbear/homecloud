import type { PageLoad } from './$types';
import { getCurrentSubdomain } from '$lib/api';

export const load: PageLoad = async () => {
	const subdomain = await getCurrentSubdomain();
	return {
		subdomain: subdomain,
	}
}
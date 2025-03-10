import type { PageLoad } from './$types';
import { getApps } from '$lib/api';

export const load: PageLoad = async () => {
	return {
		apps: await getApps()
	}
}
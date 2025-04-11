import type { PageLoad } from './$types';
import { getAvailableUpdates, updatePackageList } from '$lib/api';

export const load: PageLoad = async () => {
	await updatePackageList();
	const response = await getAvailableUpdates();
	return {
		updates: response
	}
}

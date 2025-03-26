import type { PageLoad } from './$types';
import { getStoreHome } from '$lib/api';

export const load: PageLoad = async () => {
	const response = await getStoreHome();
	return {
		storeHome: response
	}
 }
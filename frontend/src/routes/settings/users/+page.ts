import type { PageLoad } from './$types';
import { getUserOptions, getUsers } from '$lib/api';

export const load: PageLoad = async () => {
	return {
		users: await getUsers(),
		userOptions: await getUserOptions(),
	}
}
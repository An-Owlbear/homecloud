import type { PageLoad } from './$types';
import { getUsers } from '$lib/api';

export const load: PageLoad = async () => {
	return {
		users: await getUsers()
	}
}
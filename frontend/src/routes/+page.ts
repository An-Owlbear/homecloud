import type { PageLoad } from './$types';
import { getUserOptions } from '$lib/api';
import { goto } from '$app/navigation';

export const load: PageLoad = async () => {
	const options = await getUserOptions();
	if (options.completed_welcome) {
		await goto('/apps', { replaceState: true });
	}
}
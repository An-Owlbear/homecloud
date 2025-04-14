import type { PageLoad } from './$types';
import { goto } from '$app/navigation';
import { getUserOptionsState } from '$lib/userOptions.svelte';

export const load: PageLoad = async () => {
	const userOptionsState = getUserOptionsState();
	await userOptionsState.loadOptions();

	if (userOptionsState.options.completed_welcome) {
		await goto('/apps', { replaceState: true });
	}
}
import type { PageLoad } from './$types';
import { getUserOptionsState } from '$lib/userOptions.svelte';

export const ssr = false;

export const load: PageLoad = async () => {
	const userOptionsState = getUserOptionsState();
	await userOptionsState.loadOptions();
}
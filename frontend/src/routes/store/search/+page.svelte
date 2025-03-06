<script lang="ts">
	import { SearchOutline } from 'flowbite-svelte-icons';
	import type { PageProps } from './$types';
	import Package from '$lib/store/package.svelte';
	import { Input } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { searchPackages } from '$lib/api';

	const { data }: PageProps = $props();
	let packages = $state([...data.packages]);
	let search = $state(page.url.searchParams.get('q') ?? '');
	const initialUrl = page.url.searchParams.get('q') ?? '';
	let searchUrl = $derived(page.url.searchParams.get('q') ?? '');

	// Updates package list when search changes
	$effect(() => {
		if (search !== initialUrl) {
			searchPackages(search).then(p => packages = p);
		} else {
			packages = [...data.packages];
		}
	})

	// Updates search when url changes
	$effect(() => {
		search = searchUrl;
	})

	// Updates url, only used when specifically set to, as to not pollute the browser history
	const updateUrl = () => {
		goto('?' + new URLSearchParams({
			q: search,
		}));
	}
</script>

<h1 class="text-5xl mb-5">Available apps</h1>
<form>
	<Input placeholder="search" size="lg" class="mb-5" bind:value={search} onblur={updateUrl}>
		<SearchOutline slot="left" class="w-6 h-6 text-gray-500 dark:text-gray-400" />
	</Input>
</form>
<ul class="space-y-3">
	{#each packages as appPackage (appPackage.id)}
		<li>
			<Package appPackage={appPackage} />
		</li>
	{/each}
</ul>
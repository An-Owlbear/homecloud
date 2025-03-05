<script lang="ts">
	import { SearchOutline } from 'flowbite-svelte-icons';
	import type { PageProps } from './$types';
	import Package from './package.svelte';
	import { Input } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { searchPackages } from '$lib/api';

	const { data }: PageProps = $props();
	let packages = $state([...data.packages])
	let search = $state(page.url.searchParams.get('q') ?? '')

	// Updates url when search changes
	$effect(() => {
		if (search) {
			searchPackages(search).then(p => packages = p);
		} else {
			packages = [...data.packages];
		}
	})

	const updateUrl = () => {
		goto('?' + new URLSearchParams({
			q: search,
		}));
	}
</script>

<div class="container mx-auto my-10">
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
</div>
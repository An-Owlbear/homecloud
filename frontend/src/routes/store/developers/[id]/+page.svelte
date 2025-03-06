<script lang="ts">
	import { Heading } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { page } from '$app/state';
	import Package from '$lib/store/package.svelte';

	const { data }: PageProps = $props();
</script>

<Heading tag="h2" class="mb-4">Apps by {page.params.id}</Heading>
<ul class="space-y-3">
	{#each data.packages as appPackage (appPackage.id)}
		<li>
			<Package appPackage={appPackage} />
		</li>
	{/each}
</ul>
<script lang="ts">
	import Package from '$lib/store/package.svelte';
	import type { PageProps } from './$types';
	import { Button, Spinner } from 'flowbite-svelte';
	import { updateAppsRequest } from '$lib/api';
	import { invalidateAll } from '$app/navigation';

	const { data }: PageProps = $props();

	let updating = $state(false);

	const updateAppsBtn = async () => {
		updating = true;
		await updateAppsRequest();
		await invalidateAll();
		updating = false;
	}

</script>

<div class="flex flex-row justify-between items-center mb-4">
	<h1 class="text-3xl font-bold">Available Updates</h1>
	{#if data.updates.length !== 0}
		<Button class="enabled:hover:cursor-pointer space-x-2" disabled={updating} onclick={updateAppsBtn}>
			{#if updating}
				<Spinner size={5} color="white" />
				<span>Updating apps, may take a while</span>
			{:else}
				<span>Update apps</span>
			{/if}
		</Button>
	{/if}
</div>
<ul class="space-y-3">
	{#each data.updates as update (update.id)}
		<Package appPackage={update} showInstallBtn={false} />
	{/each}
</ul>
{#if data.updates.length === 0}
	<h2 class="text-xl font-semibold">No updates available</h2>
{/if}
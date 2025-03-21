<script lang="ts">
	import { Button, Card, Spinner } from 'flowbite-svelte';
	import { TrashBinOutline } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';
	import { uninstallApp } from '$lib/api';

	const { app, onUninstall }: {
		app: HomecloudApp,
		onUninstall: (app: HomecloudApp) => void,
	} = $props();


	let loading = $state(false);
	let status = $state('');

	const uninstall = async (app: HomecloudApp) => {
		loading = true;
		status = "Uninstalling";
		await uninstallApp(app.id);
		loading = false;
		status = '';
		onUninstall(app);
	}</script>

<Card horizontal class="max-w-none flex flex-row gap-10 has-[.cancel-hover:hover]:bg-white dark:has-[.cancel-hover:hover]:bg-gray-800">
	<img src="{app.image_url}" class="w-20 h-20" alt="icon for {app.name}" />
	<div class="flex flex-col">
		<span class="text-xl">{app.name}</span>
		<span class="text-md">{app.author}</span>
	</div>
	<div class="grow"></div>
	<Button class="hover:cursor-pointer self-center cancel-hover" href="/apps/{app.id}/backup">
		<span class="text-lg">Backup app data</span>
	</Button>
	<Button class={['self-center', 'space-x-2', 'cancel-hover', !loading && 'hover:cursor-pointer']} onclick={() => uninstall(app)}>
		{#if loading}
			<Spinner size={5} color="white" />
			<span class="text-lg">{status}</span>
		{:else}
			<TrashBinOutline />
			<span class="text-lg">Uninstall</span>
		{/if}
	</Button>
</Card>
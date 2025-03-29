<script lang="ts">
	import { Button, Card, Modal, Spinner } from 'flowbite-svelte';
	import { TrashBinOutline } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';
	import { uninstallApp } from '$lib/api';

	const { app, onUninstall }: {
		app: HomecloudApp,
		onUninstall: (app: HomecloudApp) => void,
	} = $props();


	let loading = $state(false);
	let status = $state('');
	let uninstallModalOpen = $state(false);

	const showUninstallModal = (event: MouseEvent) => {
		event.preventDefault();
		uninstallModalOpen = true;
	}

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
	<Button class="hover:cursor-pointer self-center cancel-hover" href="/apps/{app.id}/restore">
		<span class="text-lg">Restore app data</span>
	</Button>
	<Button class={['self-center', 'space-x-2', 'cancel-hover', !loading && 'hover:cursor-pointer']} onclick={showUninstallModal}>
		{#if loading}
			<Spinner size={5} color="white" />
			<span class="text-lg">{status}</span>
		{:else}
			<TrashBinOutline />
			<span class="text-lg">Uninstall</span>
		{/if}
	</Button>
</Card>

<Modal bind:open={uninstallModalOpen} size="sm" autoclose>
	<h2 class="mb-4 text-xl font-semibold">Are you sure you want to uninstall {app.name}</h2>
	<p class="mb-4 text-md">All app data will be removed</p>
	<div class="flex flex-row justify-between items-center gap-3">
		<Button color="alternative" class="hover:cursor-pointer">Cancel</Button>
		<Button color="red" onclick={() => uninstall(app)} class="hover:cursor-pointer">Uninstall</Button>
	</div>
</Modal>

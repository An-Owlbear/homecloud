<script lang="ts">
	import { Button, Card, Dropdown, DropdownItem, Spinner } from 'flowbite-svelte';
	import { DotsVerticalOutline } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';
	import { CheckAuthRedirect } from '$lib/api';

	const { app, onUninstall }: {
		app: HomecloudApp
		onUninstall: (app: HomecloudApp) => void
	} = $props();

	let loading = $state(false);
	let loadingMessage = $state('');
	let dropdown = $state(false);

	const uninstall = async (event: MouseEvent) => {
		event.preventDefault();
		dropdown = false;
		loading = true;
		loadingMessage = 'Uninstalling';
		const response = await fetch(`/api/v1/apps/${app.id}/uninstall`, {method: 'POST'});
		if (!response.ok) {
			await CheckAuthRedirect(response);
		}
		onUninstall(app);
	}

	const stop = async (event: MouseEvent) => {
		event.preventDefault();
		dropdown = false;
		loading = true;
		loadingMessage = 'Stopping'
		const response = await fetch(`/api/v1/apps/${app.id}/stop`, {method: 'POST'});
		if (!response.ok) {
			await CheckAuthRedirect(response);
		}
		loading = false;
	}

	const start = async (event: MouseEvent) => {
		event.preventDefault();
		dropdown = false;
		loading = true;
		loadingMessage = 'Starting';
		const response = await fetch(`/api/v1/apps/${app.id}/start`, {method: 'POST'});
		if (!response.ok) {
			await CheckAuthRedirect(response);
		}
		loading = false;
	}
</script>

<Card class="w-42 h-full overflow-hidden space-y-4 has-[.app-options:hover]:bg-white dark:has-[.app-options:hover]:bg-gray-800" href={loading ? undefined : `http://${app.name}.hc.anowlbear.com:1323`} target="_blank">
	{#if loading}
		<Spinner size="xl" />
		<span class="text-md text-center">{loadingMessage} {app.name}</span>
	{:else}
		<img src="/assets/data/{app.id}/icon.png" alt="Logo for {app.name}" class="w-30 h-30" />
		<div class="flex flex-row justify-between items-center">
			<span class="text-md truncate">{app.name}</span>
			<Button pill color="alternative" class="app-options px-1 py-1 hover:cursor-pointer" on:click={(event) => event.preventDefault()}>
				<DotsVerticalOutline size="md" />
			</Button>
			<Dropdown bind:open={dropdown}>
				<DropdownItem class="app-options" href="http://{app.name}.hc.anowlbear.com:1323" target="_blank">Open app</DropdownItem>
				<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={start}>Start app</DropdownItem>
				<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={stop}>Stop app</DropdownItem>
				<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={uninstall}>Uninstall app</DropdownItem>
			</Dropdown>
		</div>
	{/if}
</Card>

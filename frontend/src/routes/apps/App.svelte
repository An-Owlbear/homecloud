<script lang="ts">
	import { Badge, Button, Card, Dropdown, DropdownItem, Modal, Spinner } from 'flowbite-svelte';
	import { DotsVerticalOutline } from 'flowbite-svelte-icons';
	import { AppStatus, type HomecloudApp, UserRoles } from '$lib/models';
	import { CheckAuthRedirect, uninstallApp } from '$lib/api';
	import { page } from '$app/state';
	import { getUserOptionsState } from '$lib/userOptions.svelte';

	const { app, onUninstall }: {
		app: HomecloudApp
		onUninstall: (app: HomecloudApp) => void
	} = $props();

	const isAdmin = $derived(getUserOptionsState().options.user_roles.includes(UserRoles.Admin));

	const appUrl = (() => {
		let pageUrl = new URL(page.url.toString());
		pageUrl.hostname = `${app.id}.${pageUrl.hostname}`;
		pageUrl.pathname = '';
		pageUrl.search = '';
		return pageUrl.toString();
	})();

	let loading = $state(false);
	let loadingMessage = $state('');
	let dropdown = $state(false);
	let uninstallModalOpen = $state(false);

	const showUninstallModal = (event: MouseEvent) => {
		event.preventDefault();
		dropdown = false;
		uninstallModalOpen = true;
	}

	const uninstall = async (event: MouseEvent) => {
		event.preventDefault();
		dropdown = false;
		loading = true;
		loadingMessage = 'Uninstalling';
		await uninstallApp(app.id)
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
		app.status = AppStatus.Exited;
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
		app.status = AppStatus.Running;
	}
</script>

<Card class="w-42 h-full overflow-hidden space-y-4 has-[.app-options:hover]:bg-white dark:has-[.app-options:hover]:bg-gray-800" href={loading || app.status !== AppStatus.Running ? undefined : appUrl} target="_blank">
	{#if loading}
		<Spinner size="xl" />
		<span class="text-md text-center">{loadingMessage} {app.name}</span>
	{:else}
		<img src={app.image_url} alt="Logo for {app.name}" class={['w-30', 'h-30', app.status === AppStatus.Exited && 'grayscale-100']} />
		<span class="text-md truncate text-center">{app.name}</span>
		<div class="flex flex-row justify-between items-center gap-3">
			{#if app.status === AppStatus.Exited}
				<Badge color="indigo" class="text-center py-2 grow">Stopped</Badge>
			{:else if app.status === AppStatus.Running}
				<Badge color="green" class="text-center py-2 grow">Running</Badge>
			{/if}
			<Button pill color="alternative" class="app-options px-1 py-1 hover:cursor-pointer" on:click={(event) => event.preventDefault()}>
				<DotsVerticalOutline size="md" />
			</Button>
			<Dropdown bind:open={dropdown}>
				{#if app.status === AppStatus.Exited}
					{#if isAdmin}
						<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={start}>Start app</DropdownItem>
					{:else}
						<DropdownItem class="app-options">No options available</DropdownItem>
					{/if}
				{:else if app.status === AppStatus.Running}
					<DropdownItem class="app-options" href={appUrl} target="_blank">Open app</DropdownItem>
					{#if isAdmin}
						<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={stop}>Stop app</DropdownItem>
					{/if}
				{/if}
				{#if isAdmin}
					<DropdownItem class="app-options hover:cursor-pointer" role="button" on:click={showUninstallModal}>Uninstall app</DropdownItem>
				{/if}
			</Dropdown>
		</div>
	{/if}
</Card>

<Modal bind:open={uninstallModalOpen} size="sm" autoclose>
	<h2 class="mb-4 text-xl font-semibold">Are you sure you want to uninstall {app.name}</h2>
	<p class="mb-4 text-md">All app data will be removed</p>
	<div class="flex flex-row justify-between items-center gap-3">
		<Button color="alternative" class="hover:cursor-pointer">Cancel</Button>
		<Button color="red" onclick={uninstall} class="hover:cursor-pointer">Uninstall</Button>
	</div>
</Modal>

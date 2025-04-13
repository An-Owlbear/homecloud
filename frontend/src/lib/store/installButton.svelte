<script lang="ts">
	import { type PackageListItem, UserRoles } from '$lib/models';
	import { installPackage } from '$lib/api';
	import { Button, Spinner } from 'flowbite-svelte';
	import { ArrowDownToBracketOutline } from 'flowbite-svelte-icons';
	import { getUserOptionsState } from '$lib/userOptions.svelte';

	const { appPackage }: {
		appPackage: PackageListItem
	} = $props();

	let status = $state('');
	let loading = $state(false);
	const userOptions = $derived(getUserOptionsState().options);
	const hasPermission = $derived(userOptions.user_roles.includes(UserRoles.Admin));
	const enabled = $derived(hasPermission && !appPackage.installed);

	const install = async (event: MouseEvent) => {
		event.preventDefault();
		if (loading) {
			return;
		}

		status = 'Installing';
		loading = true;
		await installPackage(appPackage.id);
		appPackage.installed = true;
		loading = false;
		status = '';
	}
</script>

<Button class={['self-center', 'space-x-2', 'cancel-hover', !loading && enabled && 'hover:cursor-pointer']} disabled={!enabled} onclick={install}>
	{#if appPackage.installed}
		<span class="text-lg">Installed</span>
	{:else if !hasPermission}
		<span class="text-lg">You don't have permission to install</span>
	{:else if loading}
		<Spinner size={5} color="white" />
		<span class="text-lg">{status}</span>
	{:else}
		<ArrowDownToBracketOutline />
		<span class="text-lg">Install</span>
	{/if}
</Button>
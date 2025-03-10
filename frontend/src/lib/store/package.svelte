<script lang="ts">
	import { Button, Card, Spinner } from 'flowbite-svelte';
	import { ArrowDownToBracketOutline } from 'flowbite-svelte-icons';
	import type { PackageListItem } from '$lib/models';
	import { CheckAuthRedirect } from '$lib/api';

	const { appPackage }: {
		appPackage: PackageListItem
	} = $props();

	let status = $state('');
	let loading = $state(false);

	const install = async (event: MouseEvent) => {
		event.preventDefault();
		if (loading) {
			return;
		}

		status = 'Installing';
		loading = true;

		const response = await fetch(`/api/v1/packages/${appPackage.id}/install`, { method: 'POST' });
		if (!response.ok) {
			await CheckAuthRedirect(response);
		}

		loading = false;
		status = '';
	}
</script>

<Card horizontal class="max-w-none flex flex-row space-x-10 has-[.cancel-hover:hover]:bg-white dark:has-[.cancel-hover:hover]:bg-gray-800" href="/store/apps/{appPackage.id}">
	<img src="{appPackage.image_url}" class="w-20 h-20" alt="icon for {appPackage.name}" />
	<div class="flex flex-col">
		<span class="text-xl">{appPackage.name}</span>
		<span class="text-md">{appPackage.author}</span>
	</div>
	<div class="grow"></div>
	<Button class={['self-center', 'space-x-2', 'cancel-hover', !loading && 'hover:cursor-pointer']} disabled={appPackage.installed} onclick={install}>
		{#if appPackage.installed}
			<span class="text-lg">Installed</span>
		{:else if loading}
			<Spinner size={5} color="white" />
			<span class="text-lg">{status}</span>
		{:else}
			<ArrowDownToBracketOutline />
			<span class="text-lg">Install</span>
		{/if}
	</Button>
</Card>

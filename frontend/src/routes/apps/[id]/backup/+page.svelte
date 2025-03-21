<script lang="ts">


	import { backupApp, listExternalStorage } from '$lib/api';
	import { Button, Radio, Skeleton, Spinner } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { afterNavigate } from '$app/navigation';
	import { ArrowLeftToBracketOutline } from 'flowbite-svelte-icons';

	let previousPage: string = $state('/apps');
	afterNavigate(({ from }) => {
		previousPage = from?.url?.pathname || previousPage;
	})

	const screens = {
		drives: 'drives',
		complete: 'complete'
	}
	let currentScreen = $state(screens.drives);

	let storageDevices = $state(listExternalStorage());
	let selectedDrive = $state<string | undefined>(undefined);
	let loading = $state(false);
	let backupButtonEnabled = $derived(!!selectedDrive && !loading);

	const refreshDevices = () => {
		selectedDrive = undefined;
		storageDevices = listExternalStorage();
	}

	const backupData = async () => {
		loading = true;
		await backupApp(selectedDrive!, page.params.id);
		currentScreen = screens.complete;
		loading = false;
	}
</script>


<div class="container mx-auto mt-5">
	<a href={previousPage} class="flex flex-row items-center gap-2 mb-4 hover:text-primary-600">
		<ArrowLeftToBracketOutline />
		<span>Back to home</span>
	</a>
	{#if currentScreen === screens.drives}
		<h1 class="text-4xl font-bold mb-5">Select a drive to backup to</h1>
		{#await storageDevices}
			<Skeleton />
		{:then storageDevices}
			{#each storageDevices as storageDevice (storageDevice.name)}
				<Radio custom value={storageDevice.name} bind:group={selectedDrive} disabled={loading}>
					<div class="w-full flex flex-col p-5 text-gray-500 bg-white rounded-lg border border-gray-200 cursor-pointer dark:hover:text-gray-300 dark:border-gray-700 dark:peer-checked:text-primary-500 peer-checked:border-primary-600 peer-checked:text-primary-600 hover:text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:bg-gray-800 dark:hover:bg-gray-700">
						<span class="text-lg font-semibold">{storageDevice.label}</span>
						<span>{(storageDevice.size / 1e9).toFixed(2)}GB</span>
					</div>
				</Radio>
			{/each}
		{/await}
		<div class="mt-3 flex flex-row justify-between">
			<Button color="alternative" class="hover:cursor-pointer" onclick={refreshDevices}>Refresh external storage</Button>
			<Button disabled={!backupButtonEnabled} class={[backupButtonEnabled && 'hover:cursor-pointer']} onclick={backupData}>
				{#if !loading}
					<span>Backup</span>
				{:else}
					<div class="flex flex-row gap-2 items-center">
						<Spinner size="5" />
						<span>Backing up data</span>
					</div>
				{/if}
			</Button>
		</div>
	{:else if currentScreen === screens.complete}
		<h1 class="text-4xl font-bold mb-5">Backup complete!</h1>
		<Button href={previousPage} class="hover:cursor-pointer">Return to previous page</Button>
	{/if}
</div>
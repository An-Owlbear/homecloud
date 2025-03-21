<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { listBackups, listExternalStorage, restoreBackup } from '$lib/api';
	import { page } from '$app/state';
	import { Button, Radio, Skeleton, Spinner } from 'flowbite-svelte';
	import RadioSelectBox from '$lib/RadioSelectBox.svelte';
	import { ArrowLeftToBracketOutline } from 'flowbite-svelte-icons';
	import { DateTime } from 'luxon';

	let previousPage: string = $state('/apps');
	afterNavigate(({ from }) => {
		previousPage = from?.url?.pathname || previousPage;
	})

	const screens = {
		drives: 'drives',
		backups: 'backups',
		complete: 'complete'
	}
	let currentScreen = $state(screens.drives);

	let storageDevices = $state(listExternalStorage());
	let selectedDrive = $state<string | undefined>(undefined);

	// Uses derived by to only run request once
	let backupRefreshTrigger = $state(1);
	let backups = $derived.by(() => {
		console.log('test');
		if (selectedDrive && backupRefreshTrigger) {
			return listBackups(selectedDrive, page.params.id);
		}
		return Promise.resolve([]);
	})
	let selectedBackup = $state<string | undefined>(undefined);

	let restoreLoading = $state(false);
	let backupButtonEnabled = $derived(!!selectedDrive && !!selectedBackup && !restoreLoading);

	const refreshDevices = () => {
		selectedDrive = undefined;
		storageDevices = listExternalStorage();
	}

	const refreshBackups = () => {
		selectedBackup = undefined;
		backupRefreshTrigger++;
	}

	const restoreDataButton = async () => {
		restoreLoading = true;
		await restoreBackup(selectedDrive!, page.params.id, selectedBackup!);
		currentScreen = screens.complete;
		restoreLoading = false;
	}
</script>

<div class="container mx-auto mt-5">
	<a href={previousPage} class="flex flex-row items-center gap-2 mb-4 hover:text-primary-600">
		<ArrowLeftToBracketOutline />
		<span>Back to home</span>
	</a>
	{#if currentScreen === screens.drives}
		<h1 class="text-4xl font-bold mb-5">Select a drive to restore from</h1>
		{#await storageDevices}
			<Skeleton />
		{:then storageDevices}
			{#each storageDevices as storageDevice (storageDevice.name)}
				<Radio custom value={storageDevice.name} bind:group={selectedDrive} disabled={restoreLoading}>
					<RadioSelectBox>
						<span class="text-lg font-semibold">{storageDevice.label}</span>
						<span>{(storageDevice.size / 1e9).toFixed(2)}GB</span>
					</RadioSelectBox>
				</Radio>
			{/each}
		{/await}
		<div class="mt-3 flex flex-row justify-between">
			<Button color="alternative" class="hover:cursor-pointer" onclick={refreshDevices}>Refresh external storage</Button>
			<Button disabled={!selectedDrive} class={[!!selectedDrive && 'hover:cursor-pointer']} onclick={() => currentScreen = screens.backups}>Next</Button>
		</div>
	{:else if currentScreen === screens.backups}
		{#await backups}
			<Skeleton />
		{:then backups}
			<h1 class="text-4xl font-bold mb-5">Select a backup to restore</h1>
			<div class="space-y-2">
				{#each backups.toReversed() as backup (backup)}
					<Radio custom value={backup} bind:group={selectedBackup} disabled={restoreLoading}>
						<RadioSelectBox>
							<span class="text-lg font-semibold">{DateTime.fromISO(backup).toLocaleString(DateTime.DATETIME_HUGE)}</span>
						</RadioSelectBox>
					</Radio>
				{/each}
			</div>
			<div class="mt-3 flex flex-row gap-2">
				<Button color="alternative" disabled={restoreLoading} class={[!restoreLoading && 'hover:cursor-pointer']} onclick={refreshBackups}>Refresh available backups</Button>
				<div class="grow"></div>
				<Button color="alternative" disabled={restoreLoading} class={[!restoreLoading && 'hover:cursor-pointer']} onclick={() => currentScreen = screens.drives}>Back</Button>
				<Button disabled={!backupButtonEnabled} class={[backupButtonEnabled && 'hover:cursor-pointer']} onclick={restoreDataButton}>
					{#if !restoreLoading}
						<span>Restore</span>
					{:else}
						<div class="flex flex-row gap-2 items-center">
							<Spinner size="5" />
							<span>Restoring data</span>
						</div>
					{/if}
				</Button>
			</div>
		{/await}
	{:else if currentScreen === screens.complete}
		<h1 class="text-4xl font-bold mb-5">Restore complete!</h1>
		<Button href={previousPage} class="hover:cursor-pointer">Return to previous page</Button>
	{/if}
</div>

<script lang="ts">
	import { Button, Heading, Spinner } from "flowbite-svelte";
	import { checkUpdates, updateSystem } from '$lib/api';
	import { CheckCircleSolid, ExclamationCircleSolid, InfoCircleSolid } from "flowbite-svelte-icons";



	let updateNeeded =  $state(checkUpdates());
	let updating = $state(false);
	let updateRequest = $state(Promise.resolve());

	const runUpdateSystem = () => {
		updating = true;
		updateRequest = updateSystem();
	}
</script>

<Heading tag="h1" class="w-auto mb-10" customSize="text-3xl font-medium">System settings</Heading>
<div class="flex flex-row items-center gap-4">
	{#await updateNeeded}
		<Spinner class="w-17 h-17" />
		<span class="text-2xl">Checking for updates</span>
	{:then updateNeeded}
		{#if updating}
			{#await updateRequest}
				<Spinner class="w-17 h-17" />
				<span class="text-2xl">Updating system, this may take a while</span>
			{:then updateRequest}
				<InfoCircleSolid class="w-20 h-20" />
				<span class="text-2xl">System updated! Now rebooting, this may take several minutes</span>
			{/await}
		{:else if updateNeeded.update_required === true}
			<ExclamationCircleSolid class="w-20 h-20 text-amber-400" />
			<span class="text-2xl">System update available</span>
			<Button class="hover:cursor-pointer" onclick={runUpdateSystem}>Update system</Button>
		{:else}
			<CheckCircleSolid class="w-20 h-20 text-green-500" />
			<span class="text-2xl">System up to date!</span>
		{/if}
	{/await}
</div>
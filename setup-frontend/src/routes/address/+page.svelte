<script lang="ts">
	import { Button, ButtonGroup, Input, InputAddon, Label, Li, List, Spinner } from 'flowbite-svelte';
	import { ArrowLeftOutline } from 'flowbite-svelte-icons';
	import ListCheck from '$lib/ListCheck.svelte';
	import { checkSubdomainTaken, registerDomain } from '$lib/api';
	import { goto } from '$app/navigation';

	let registerDomainLoading = $state(false);

	let subdomain = $state('');
	let domainUnique = $derived.by(async () => {
		if (!subdomain) {
			return false;
		}

		return !(await checkSubdomainTaken(`${subdomain}.homecloudapp.com`));
	});
	let validChars = $derived(/^[a-zA-Z0-9\-]+$/.test(subdomain));
	let validSize = $derived(!!subdomain && subdomain.length <= 20);
	let validSubmit = $derived(domainUnique && validChars && validSize);
	let displayInputError = $derived(!!subdomain && !registerDomainLoading && !validSubmit);

	const registerDomainBtn = async () => {
		registerDomainLoading = true;
		await registerDomain(subdomain);
		registerDomainLoading = false;
		await goto('/');
	}
</script>

<div class="max-w-2xl p-5 mx-auto">
	<a class="mb-4 flex flex-row items-center" href="/">
		<ArrowLeftOutline size="lg" class="me-2" />
		<p>Back</p>
	</a>
	<h1 class="text-3xl font-bold mb-4">Configure server address</h1>
	<p class="mb-4">To access your homecloud server outside your home network you'll need to have an address registered.</p>
	<p>This must:</p>
	<List class="mb-4" list="none">
		<Li icon>
			{#await domainUnique}
				<ListCheck state="loading">Be unique</ListCheck>
			{:then domainUnique}
				<ListCheck state={domainUnique ? 'passed' : 'failed'}>Be unique</ListCheck>
			{/await}
		</Li>
		<Li icon>
			<ListCheck state={validChars ? 'passed' : 'failed'}>Contain only letters, number and dashes (-)</ListCheck>
		</Li>
		<Li icon>
			<ListCheck state={validSize ? 'passed' : 'failed'}>Be no more than 20 characters long</ListCheck>
		</Li>
	</List>
	<div>
		<Label class="mb-2" for="chosen-address">Address to access your server</Label>
		<ButtonGroup class="w-full mb-4" size="md">
			<Input id="chosen-address" placeholder="Server address" bind:value={subdomain} color={displayInputError ? 'red' : 'base'} disabled={registerDomainLoading} />
			<InputAddon>.homecloudapp.com</InputAddon>
		</ButtonGroup>
		<Button class={['w-full flex flex-row items-center gap-2', validSubmit && 'cursor-pointer']} disabled={!validSubmit} onclick={registerDomainBtn}>
			{#if registerDomainLoading}
				<Spinner size="5" />
				<span>Registering domain</span>
			{:else}
				Register address
			{/if}
		</Button>
	</div>
</div>
<script lang="ts">
	import type { PageProps } from './$types';
	import { Button } from 'flowbite-svelte';

	const { data }: PageProps = $props();
</script>

<div class="max-w-2xl p-5 mx-auto space-y-4">
	<h1 class="text-3xl font-bold">Homecloud Setup</h1>
	{#if !data.subdomain}
		<p>Before using Homecloud you need to register an address to access it from</p>
	{:else}
		<p>Your Homecloud server is setup and ready to access! You can change the domain from the link above if needed</p>
	{/if}
	<Button class="w-full" href="/address">Configure address</Button>
	{#if data.subdomain}
		<Button class="w-full" href="https://{data.subdomain}.homecloudapp.com">Open Homecloud</Button>
	{/if}
</div>
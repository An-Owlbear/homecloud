<script lang="ts">
	import type { PageProps } from './$types';
	import { Button, Heading } from 'flowbite-svelte';
	import type { HomecloudApp } from '$lib/models';
	import App from './app.svelte';

	const { data }: PageProps = $props();
	let apps = $state([...data.apps]);

	const onUninstall = (app: HomecloudApp) => {
		const appIndex = apps.findIndex(a => a.id === app.id);
		if (appIndex > -1) {
			apps.splice(appIndex, 1);
		}
	}
</script>

<div class="flex flex-row justify-between mb-4">
	<Heading tag="h1" class="w-auto" customSize="text-3xl font-medium">App settings</Heading>
	<Button class="hover:cursor-pointer" href="/store/updates">Check for updates in store</Button>
</div>
<ul class="space-y-3">
	{#each apps as app (app.id)}
		<li>
			<App app={app} onUninstall={onUninstall} />
		</li>
	{/each}
</ul>
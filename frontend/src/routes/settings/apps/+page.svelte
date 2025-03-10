<script lang="ts">
	import type { PageProps } from './$types';
	import { Heading } from 'flowbite-svelte';
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

<Heading tag="h1" class="w-auto mb-4" customSize="text-3xl font-medium">App settings</Heading>
<ul class="space-y-3">
	{#each apps as app (app.id)}
		<li>
			<App app={app} onUninstall={onUninstall} />
		</li>
	{/each}
</ul>
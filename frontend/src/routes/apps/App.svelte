<script lang="ts">
	import { Card } from 'flowbite-svelte';
	import { ArchiveSolid } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';

	const { app }: {
		app: HomecloudApp
	} = $props();
</script>

<Card class="space-y-4">
	<a href="http://{app.name}.hc.anowlbear.com:1323" target="_blank">
		<ArchiveSolid class="w-31.25 h-31.25" />
		<h1>{app.name}</h1>
	</a>
</Card>
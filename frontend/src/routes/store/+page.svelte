<script lang="ts">
	import type { PageProps } from './$types';
	import Package from './package.svelte';

	const { data }: PageProps = $props();
	const packages = $state([...data.packages])
</script>

<div class="container mx-auto my-10">
	<h1 class="text-5xl mb-5">Available apps</h1>
	<ul class="space-y-3">
		{#each packages as appPackage (appPackage.id)}
			<li>
				<Package appPackage={appPackage} />
			</li>
		{/each}
	</ul>
</div>
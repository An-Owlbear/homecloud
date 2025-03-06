<script lang="ts">
	import type { PageProps } from "./$types";
	import { Heading } from "flowbite-svelte";
	import { page } from '$app/state';
	import Package from "$lib/store/package.svelte";

	const { data }: PageProps = $props();
</script>

<Heading tag="h2" class="capitalize mb-4">{page.params.category}</Heading>
<ul class="space-y-3">
	{#each data.packages as appPackage (appPackage.id)}
		<li>
			<Package appPackage={appPackage} />
		</li>
	{/each}
</ul>
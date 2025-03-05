<script lang="ts">
	import type { PageProps } from './$types'
	import { Card, Heading } from 'flowbite-svelte';

	const { data }: PageProps = $props();
</script>

<div class="container mx-auto my-10">
	<Heading tag="h2" class="mb-4">Categories</Heading>
	<ul class="grid grid-cols-2 gap-4">
		{#each data.categories as category}
			<li class="w-full">
				<Card class="max-w-none capitalize" horizontal href="/store/categories/{category}">{category}</Card>
			</li>
		{/each}
	</ul>
</div>
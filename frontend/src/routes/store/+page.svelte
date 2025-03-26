<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageProps } from './$types';
	import { Card } from 'flowbite-svelte';

	const { data }: PageProps = $props();
</script>

<div class="container">
	<h1 class="text-3xl font-bold mb-4">Popular categories</h1>
	<ul class="grid grid-cols-2 gap-4">
		{#each data.storeHome.popular_categories as category}
			<li class="w-full">
				<Card class="max-w-none capitalize" horizontal href="/store/categories/{category}">{category}</Card>
			</li>
		{/each}
	</ul>
	<h1 class="text-3xl font-bold my-4">New apps</h1>
	<ul class="flex flex-wrap flex-row space-x-8">
		{#each data.storeHome.new_apps as app}
			<Card class="w-42 h-full overflow-hidden space-y-4" href="/store/apps/{app.id}">
				<img src={app.image_url} alt="Logo for {app.name}" class="w-30 h-30" />
				<span class="text-md truncate text-center">{app.name}</span>
			</Card>
		{/each}
	</ul>
</div>
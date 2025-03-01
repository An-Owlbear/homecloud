<script lang="ts">
	import type { PageProps } from './$types';
	import App from './App.svelte';

	const { data }: PageProps = $props();
</script>

<style>
	ul {
		list-style-type: none;
	}

	.app-list {
		display: flex;
		flex-wrap: wrap;
		flex-direction: row;
	}
</style>

<div class="w-full px-10 py-8 mx-auto">
	<ul class="flex flex-wrap flex-row space-x-8">
		{#each data.apps as app}
			<li class="app">
				<App app={app} />
			</li>
		{/each}
	</ul>
</div>
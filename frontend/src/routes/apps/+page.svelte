<script lang="ts">
	import type { PageProps } from '../../../.svelte-kit/types/src/routes';

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

	.app {
		display: flex;
	}
</style>

<ul class="app-list">
	{#each data.apps as app}
		<li class="app">{app.name}</li>
	{/each}
</ul>
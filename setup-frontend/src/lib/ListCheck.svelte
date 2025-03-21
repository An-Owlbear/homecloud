<script lang="ts">
	import { CheckCircleSolid, CloseCircleSolid } from 'flowbite-svelte-icons';
	import type { Snippet } from 'svelte';

	const { passed, size = 5, children }: {
		passed: boolean,
		size?: number,
		children: Snippet
	} = $props();
</script>

<div class="flex flex-row items-center">
	{#if passed}
		<CheckCircleSolid class="w-{size} h-{size} me-2 text-green-500 dark:text-green-400" />
	{:else}
		<CloseCircleSolid class="w-{size} h-{size} me-2 text-gray-500 dark:text-gray-400" />
	{/if}
	{@render children()}
</div>
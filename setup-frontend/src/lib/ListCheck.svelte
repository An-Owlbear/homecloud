<script lang="ts">
	import { CheckCircleSolid, CloseCircleSolid } from 'flowbite-svelte-icons';
	import type { Snippet } from 'svelte';
	import { Spinner } from 'flowbite-svelte';

	const { state, size = 5, children }: {
		state: 'passed' | 'failed' | 'loading';
		size?: number,
		children: Snippet
	} = $props();
</script>

<div class="flex flex-row items-center">
	{#if state === 'passed'}
		<CheckCircleSolid class="w-{size} h-{size} me-2 text-green-500 dark:text-green-400" />
	{:else if state === 'failed'}
		<CloseCircleSolid class="w-{size} h-{size} me-2 text-gray-500 dark:text-gray-400" />
	{:else if state === 'loading'}
		<Spinner class="me-2" size={size} />
	{/if}
	{@render children()}
</div>
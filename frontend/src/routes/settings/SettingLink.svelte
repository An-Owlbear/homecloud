<script lang="ts">
	import { Card } from 'flowbite-svelte';
	import type { Component, Snippet } from 'svelte';

	const { href, Icon, title }: {
		href: string;
		Icon: Component;
		title: string;
	} = $props();
</script>

<Card horizontal class="max-w-none space-x-10 flex flex-row items-center" href={href}>
	<Icon class="w-15 h-15" />
	<span class="text-2xl font-semibold">{title}</span>
</Card>
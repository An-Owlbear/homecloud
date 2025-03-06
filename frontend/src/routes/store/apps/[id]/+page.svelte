<script lang="ts">
	import { Button, Heading, P, Span } from 'flowbite-svelte';
import type { PageProps } from './$types';
	import { ArrowDownToBracketOutline } from 'flowbite-svelte-icons';

	const { data }: PageProps = $props();
</script>

<div class="flex flex-row items-center gap-6 mb-4">
	<img src={data.package.image_url} alt="Icon for {data.package.name}" class="w-25 h-25" />
	<div>
		<Heading tag="h1" class="text-5xl font-medium">{data.package.name}</Heading>
		<Span class="text-xl font-normal">{data.package.author}</Span>
	</div>
	<div class="grow"></div>
	<Button class="space-x-2">
		<ArrowDownToBracketOutline size="lg" />
		<span class="text-lg">Install</span>
	</Button>
</div>
<Heading tag="h2" class="mb-4" customSize="text-3xl font-medium">Categories</Heading>
<ul class="mb-4">
	{#each data.package.categories as category}
		<Button pill color="light" href="/store/categories/{category}" class="capitalize">{category}</Button>
	{/each}
</ul>
<Heading tag="h2" class="mb-4" customSize="text-3xl font-medium">Description</Heading>
<P>{data.package.description}</P>
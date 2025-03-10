<script lang="ts">
	import { Button, Heading, P, Span } from 'flowbite-svelte';
import type { PageProps } from './$types';
	import InstallButton from '$lib/store/installButton.svelte';

	const { data }: PageProps = $props();
	const appPackage = $state(data.package)
</script>

<div class="flex flex-row items-center gap-6 mb-4">
	<img src={appPackage.image_url} alt="Icon for {appPackage.name}" class="w-25 h-25" />
	<div>
		<Heading tag="h1" class="text-5xl font-medium">{appPackage.name}</Heading>
		<Span class="text-xl font-normal">{appPackage.author}</Span>
	</div>
	<div class="grow"></div>
	<InstallButton appPackage={appPackage} />
</div>
<Heading tag="h2" class="mb-4" customSize="text-3xl font-medium">Categories</Heading>
<ul class="flex flex-row gap-2 mb-4">
	{#each appPackage.categories as category}
		<Button pill color="light" href="/store/categories/{category}" class="capitalize">{category}</Button>
	{/each}
</ul>
<Heading tag="h2" class="mb-4" customSize="text-3xl font-medium">Description</Heading>
<P>{appPackage.description}</P>
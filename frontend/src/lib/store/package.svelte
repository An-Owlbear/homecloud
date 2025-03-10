<script lang="ts">
	import { Card } from 'flowbite-svelte';
	import type { PackageListItem } from '$lib/models';
	import InstallButton from './installButton.svelte';

	const { appPackage }: {
		appPackage: PackageListItem
	} = $props();
</script>

<Card horizontal class="max-w-none flex flex-row space-x-10 has-[.cancel-hover:hover]:bg-white dark:has-[.cancel-hover:hover]:bg-gray-800" href="/store/apps/{appPackage.id}">
	<img src="{appPackage.image_url}" class="w-20 h-20" alt="icon for {appPackage.name}" />
	<div class="flex flex-col">
		<span class="text-xl">{appPackage.name}</span>
		<span class="text-md">{appPackage.author}</span>
	</div>
	<div class="grow"></div>
	<InstallButton appPackage={appPackage} />
</Card>

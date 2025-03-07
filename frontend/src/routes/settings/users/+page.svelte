<script lang="ts">
	import { Badge, Button, Card, Heading } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { UserCircleOutline } from 'flowbite-svelte-icons';

	const { data }: PageProps = $props();
</script>

<Heading tag="h1" class="mb-4" customSize="text-3xl font-medium">User settings</Heading>
<ul class="space-y-4">
	{#each data.users as user (user.id)}
		<Card horizontal class="max-w-none flex flex-row items-center gap-4">
			<UserCircleOutline size="xl" />
			<div class="flex flex-col">
				<span class="text-2xl">{user.traits.name}</span>
				<span class="text-xl">{user.traits.email}</span>
			</div>
			{#if (user.metadata_public?.roles?.includes("admin"))}
				<Badge color="purple" class="text-xl">Admin</Badge>
			{/if}
			<div class="grow"></div>
			<Button color="red" class="text-lg">Delete user</Button>
		</Card>
	{/each}
</ul>
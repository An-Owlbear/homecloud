<script lang="ts">
	import { Heading } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import type { User } from '$lib/models';
	import UserCard from './userCard.svelte';

	const { data }: PageProps = $props();
	let users = $state([...data.users]);

	const onDelete = (user: User) => {
		const index = users.findIndex(u => u.id === user.id);
		if (index > -1) {
			users.splice(index, 1);
		}
	}
</script>

<Heading tag="h1" class="mb-4" customSize="text-3xl font-medium">User settings</Heading>
<ul class="space-y-4">
	{#each users as user (user.id)}
		<UserCard user={user} onDelete={onDelete} />
	{/each}
</ul>
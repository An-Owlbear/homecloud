<script lang="ts">
	import type { User } from '$lib/models';
	import { Badge, Button, Card, Spinner } from 'flowbite-svelte';
	import { UserCircleOutline } from 'flowbite-svelte-icons';
	import { deleteUser } from '$lib/api';

	const { user, onDelete, currentUser, onRecover }: {
		user: User,
		onDelete: (user: User) => void,
		currentUser: boolean,
		onRecover: (user: User) => void,
	} = $props();
	let loading = $state(false);

	const deleteUserFunc = async (event: MouseEvent) => {
		event.preventDefault();
		await deleteUser(user.id);
		onDelete(user);
		loading = false;
	}
</script>

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
	{#if !currentUser}
		<Button color="purple" class="text-lg hover:cursor-pointer" onclick={() => onRecover(user)}>Recover account</Button>
		<Button color="red" class={['text-lg', !loading && 'hover:cursor-pointer']} onclick={deleteUserFunc}>
			{#if loading}
				<Spinner size={5} color="white" />
				<span>Deleting user</span>
			{:else}
				<span>Delete user</span>
			{/if}
		</Button>
	{/if}
</Card>
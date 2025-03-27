<script lang="ts">
	import { Button, Heading, Modal, Skeleton, Toast } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import type { User } from '$lib/models';
	import UserCard from './userCard.svelte';
	import { inviteUser } from '$lib/api';
	import { page } from '$app/state';
	import { FileCopySolid } from 'flowbite-svelte-icons';
	import { slide } from 'svelte/transition';
	import { toasts } from '$lib/toasts.svelte';

	const { data }: PageProps = $props();
	let users = $state([...data.users]);
	let modalOpen = $state(false);
	let inviteLink = $state<Promise<string>>(Promise.resolve(''));

	const onDelete = (user: User) => {
		const index = users.findIndex(u => u.id === user.id);
		if (index > -1) {
			users.splice(index, 1);
		}
	}

	const invite = (event: MouseEvent) => {
		event.preventDefault();
		modalOpen = true;

		inviteLink = (async () => {
			const inviteCode = await inviteUser()
			return `${page.url.protocol}//${page.url.host}/auth/registration?code=${inviteCode.code}`;
		})();
	}

	const copyInvite = async (event: MouseEvent) => {
		await navigator.clipboard.writeText(await inviteLink);
		toasts.push({ content: 'Invite link copied!' });
	}
</script>

<div class="mb-4 flex flex-row justify-between items-center">
	<Heading tag="h1" class="w-auto" customSize="text-3xl font-medium">User settings</Heading>
	<Button class="text-lg hover:cursor-pointer" onclick={invite}>Invite new user</Button>
</div>
<ul class="space-y-4">
	{#each users as user (user.id)}
		<UserCard user={user} onDelete={onDelete} currentUser={user.id === data.userOptions.user_id} />
	{/each}
</ul>

<Modal title="Invite new user" bind:open={modalOpen} outsideclose>
	{#await inviteLink}
		<span class="block text-md">Creating new invite code</span>
		<Skeleton />
	{:then inviteLink}
		<span class="block text-md">You can invite a new user by copying and sending them the link below. The link can only invite one user and will expire afterwards.</span>
		<span class="block text-md">Click the link below to copy</span>
		<Button class="w-full space-x-2 hover:cursor-pointer" color="alternative" onclick={copyInvite}>
			<FileCopySolid size="lg" />
			<span class="text-md truncate">{inviteLink}</span>
		</Button>
	{/await}
</Modal>
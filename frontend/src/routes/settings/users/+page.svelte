<script lang="ts">
	import { Button, Heading, Modal, Skeleton, Toast } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import type { RecoveryCode, User } from '$lib/models';
	import UserCard from './userCard.svelte';
	import { createRecoveryCode, inviteUser } from '$lib/api';
	import { page } from '$app/state';
	import { FileCopySolid } from 'flowbite-svelte-icons';
	import { slide } from 'svelte/transition';
	import { toasts } from '$lib/toasts.svelte';

	const { data }: PageProps = $props();
	let users = $state([...data.users]);
	let inviteModalOpen = $state(false);
	let inviteLink = $state<Promise<string>>(Promise.resolve(''));

	const onDelete = (user: User) => {
		const index = users.findIndex(u => u.id === user.id);
		if (index > -1) {
			users.splice(index, 1);
		}
	}

	const invite = (event: MouseEvent) => {
		event.preventDefault();
		inviteModalOpen = true;

		inviteLink = (async () => {
			const inviteCode = await inviteUser()
			return `${page.url.protocol}//${page.url.host}/auth/registration?code=${inviteCode.code}`;
		})();
	}

	const copyInvite = async (event: MouseEvent) => {
		await navigator.clipboard.writeText(await inviteLink);
		toasts.push({ content: 'Invite link copied!' });
	}

	let recoveryModalOpen = $state(false);
	let recoveryCode = $state<Promise<RecoveryCode>>();

	const recoverUser = (user: User) => {
		recoveryModalOpen = true;
		recoveryCode = createRecoveryCode(user.id);
	}

	const copyRecoveryLink = async () => {
		if (recoveryCode !== undefined) {
			await navigator.clipboard.writeText((await recoveryCode).recovery_link);
		}
		toasts.push({ content: 'Recovery link copied!' });
	}
</script>

<div class="mb-4 flex flex-row justify-between items-center">
	<Heading tag="h1" class="w-auto" customSize="text-3xl font-medium">User settings</Heading>
	<Button class="text-lg hover:cursor-pointer" onclick={invite}>Invite new user</Button>
</div>
<ul class="space-y-4">
	{#each users as user (user.id)}
		<UserCard user={user} onDelete={onDelete} currentUser={user.id === data.userOptions.user_id} onRecover={recoverUser} />
	{/each}
</ul>

<Modal title="Invite new user" bind:open={inviteModalOpen} outsideclose>
	{#await inviteLink}
		<span class="block text-md">Creating new invite code</span>
		<Skeleton />
	{:then inviteLink}
		<span class="block text-md">You can invite a new user by copying the below and sending it to them with an E-Mail or message. They can then open this link to create their account. The link can only invite one user and will expire afterwards, or if it isn't used within 24 hours.</span>
		<span class="block text-md">Click the link below to copy</span>
		<Button class="w-full space-x-2 hover:cursor-pointer" color="alternative" onclick={copyInvite}>
			<FileCopySolid size="lg" />
			<span class="text-md truncate">{inviteLink}</span>
		</Button>
	{/await}
</Modal>

<Modal title="Recover user" bind:open={recoveryModalOpen} outsideclose>
	{#await recoveryCode}
		<span class="block text-md">Creating recovery code</span>
		<Skeleton />
	{:then recoveryCode}
		<span class="block text-md">The user's account can be recovered using the link below, you will need to send it to them through an E-Mail or another messaging app.</span>
		<span class="block text-md">Click the link below to copy</span>
		<Button class="w-full space-x-2 hover:cursor-pointer" color="alternative" onclick={copyRecoveryLink}>
			<FileCopySolid size="lg" />
			<span class="text-md truncate">{recoveryCode?.recovery_link}</span>
		</Button>
		<span class="block text-md">They will also need the verification code below, copy and send it to them along with the link.</span>
		<span class="block text-2xl">{recoveryCode?.recovery_code}</span>
	{/await}
</Modal>
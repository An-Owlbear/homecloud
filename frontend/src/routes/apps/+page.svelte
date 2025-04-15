<script lang="ts">
	import { Button, Dropdown, DropdownItem, Navbar, NavBrand, NavHamburger, NavLi, NavUl } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import App from './App.svelte';
	import { ChevronDownOutline, CogSolid, ShoppingBagSolid, UserCircleSolid } from 'flowbite-svelte-icons';
	import { type HomecloudApp, UserRoles } from '$lib/models';
	import { getUserOptionsState } from '$lib/userOptions.svelte';

	const { data }: PageProps = $props();
	const apps = $state([...data.apps]);
	const userOptions = $derived(getUserOptionsState().options);

	const onUninstall = (app: HomecloudApp) => {
		const index = apps.findIndex(a => a.id === app.id);
		if (index > -1) {
			apps.splice(index, 1);
		}
	};
</script>

<Navbar>
	<NavBrand href="/" class="text-gray-700">
		<span class="self-center whitespace-nowrap text-xl font-semibold">Homecloud</span>
	</NavBrand>
	<NavHamburger />
	<NavUl>
		<NavLi href="/store" class="flex flex-row items-center space-x-2">
			<ShoppingBagSolid />
			<span>App store</span>
		</NavLi>
		{#if userOptions.user_roles.includes(UserRoles.Admin)}
			<NavLi href="/settings" class="flex flex-row items-center space-x-2">
				<CogSolid />
				<span>Settings</span>
			</NavLi>
		{/if}
		<NavLi class="flex flex-col">
			<div class="flex flex-row items-center gap-2 hover:cursor-pointer">
				<UserCircleSolid />
				<span>My Account</span>
				<ChevronDownOutline />
			</div>
			<Dropdown>
				<DropdownItem href="/auth/settings">My Account Settings</DropdownItem>
				<DropdownItem class="p-0">
					<form id="homecloud-logout-form" method="post" action="/auth/logout">
						<button type="submit" class="hover:cursor-pointer w-full h-full px-4 py-2 text-left">Logout</button>
					</form>
				</DropdownItem>
			</Dropdown>
		</NavLi>
	</NavUl>
</Navbar>
<div class="container mx-auto my-10">
	{#if apps.length === 0}
		<h1 class="text-3xl font-bold">No apps installed</h1>
	{/if}
	<ul class="flex flex-wrap flex-row space-x-8">
		{#each apps as app (app.id)}
			<li class="app">
				<App app={app} onUninstall={onUninstall} />
			</li>
		{/each}
	</ul>
</div>
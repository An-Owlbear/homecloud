<script lang="ts">
	import { Navbar, NavBrand, NavHamburger, NavLi, NavUl } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import App from './App.svelte';
	import { CogSolid, ShoppingBagSolid, UserCircleSolid } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';

	const { data }: PageProps = $props();
	const apps = $state([...data.apps]);

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
		<NavLi href="/settings" class="flex flex-row items-center space-x-2">
			<CogSolid />
			<span>Settings</span>
		</NavLi>
		<NavLi href="/user" class="flex flex-row items-center space-x-2">
			<UserCircleSolid />
			<span>My Account</span>
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
<script lang="ts">
	import { Navbar, NavBrand, NavHamburger, NavLi, NavUl } from 'flowbite-svelte';
  import type { PageProps } from './$types';
	import App from './App.svelte';
	import { CogSolid, UserCircleSolid } from 'flowbite-svelte-icons';
	import type { HomecloudApp } from '$lib/models';

	const { data }: PageProps = $props();

	const onUninstall = (app: HomecloudApp) => {
		const index = data.apps.findIndex(app => app.id === app.id);
		if (index > -1) {
			data.apps.splice(index, 1);
		}
	}
</script>

<Navbar>
	<NavBrand href="/" class="text-gray-700">
		<span class="self-center whitespace-nowrap text-xl font-semibold">Homecloud</span>
	</NavBrand>
	<NavHamburger />
	<NavUl>
		<NavLi href="/settings" class="flex flex-row items-center space-x-2">
			<CogSolid />
			<span>Settings</span>
		</NavLi>
		<NavLi href="/user" class="flex flex-row items-center space-x-2">
			<UserCircleSolid />
			<span>Account</span>
		</NavLi>
	</NavUl>
</Navbar>
<div class="container mx-auto my-10">
	<ul class="flex flex-wrap flex-row space-x-8">
		{#each data.apps as app}
			<li class="app">
				<App app={app} onUninstall={onUninstall} />
			</li>
		{/each}
	</ul>
</div>
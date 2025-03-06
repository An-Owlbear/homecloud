<script lang="ts">
	import { Navbar, NavBrand, NavHamburger, NavLi, NavUl } from 'flowbite-svelte';
	import { ArchiveSolid, ArrowLeftToBracketOutline, CogSolid, InfoCircleSolid, UsersSolid } from 'flowbite-svelte-icons';

	let { children } = $props();
</script>

<Navbar>
	<NavBrand href="/settings" class="text-gray-700 gap-2">
		<CogSolid />
		<span class="self-center whitespace-nowrap text-xl font-semibold">Settings</span>
	</NavBrand>
	<NavHamburger />
	<NavUl>
		<NavLi href="/settings/users" class="flex flex-row items-center gap-2">
			<UsersSolid />
			<span>User settings</span>
		</NavLi>
		<NavLi href="/settings/apps" class="flex flex-row items-center gap-2">
			<ArchiveSolid />
			<span>App settings</span>
		</NavLi>
		<NavLi href="/settings/system" class="flex flex-row items-center gap-2">
			<InfoCircleSolid />
			<span>System settings</span>
		</NavLi>
		<NavLi href="/apps" class="flex flex-row items-center gap-2">
			<ArrowLeftToBracketOutline />
			<span>Back to home</span>
		</NavLi>
	</NavUl>
</Navbar>

<div class="container mx-auto my-4">
	{@render children()}
</div>

<script lang="ts">
	import { GridSolid, HomeSolid, ShoppingBagSolid, UsersSolid } from "flowbite-svelte-icons";
	import { Button } from 'flowbite-svelte';
	import { goto } from '$app/navigation';

	const screens = [
		'welcome',
		'store',
		'apps',
		'users',
		'start'
	];

	let currentScreen = $state(screens[0]);

	const nextScreen = () => {
		const nextIndex = screens.indexOf(currentScreen) + 1;
		if (nextIndex < screens.length) currentScreen = screens[nextIndex];
		else goto('/apps');
	}

	const previousScreen = () => {
		const previousIndex = screens.indexOf(currentScreen) - 1;
		if (previousIndex >= 0) currentScreen = screens[previousIndex];
	}

</script>

<div class="max-w-2xl mx-auto my-10 space-y-4 flex flex-col items-center text-center">
	{#if currentScreen === screens[0]}
		<h1 class="text-5xl font-bold">Welcome to Homecloud!</h1>
		<HomeSolid class="w-25 h-25" />
		<p class="text-xl">Homecloud lets you easily install and access your own private online apps</p>
	{:else if currentScreen === screens[1]}
		<h1 class="text-5xl font-bold">Installing new apps</h1>
		<ShoppingBagSolid class="w-25 h-25" />
		<p class="text-xl">Open the app store to install new apps to your server</p>
	{:else if currentScreen === screens[2]}
		<h1 class="text-5xl font-bold">Opening apps</h1>
		<GridSolid class="w-25 h-25" />
		<p class="text-xl">Your apps can be opened a single click from the homepage</p>
	{:else if currentScreen === screens[3]}
		<h1 class="text-5xl font-bold">Multiple users</h1>
		<UsersSolid class="w-25 h-25" />
		<p class="text-xl">You can easily invite friends and family to use and your online apps</p>
		<p class="text-xl">These users won't have permission to install new apps so you won't have to worry about someone else making changes</p>
	{:else if currentScreen === screens[4]}
		<h1 class="text-5xl font-bold">Start</h1>
		<p class="text-xl">That's all you need to know to use Homecloud, click the 'Next' button below to start!</p>
	{/if}
	<div class="flex flex-row w-full mt-10 px-10 justify-between">
		<Button disabled={currentScreen === screens[0]} class="enabled:cursor-pointer" onclick={previousScreen}>Go back</Button>
		<Button class="enabled:cursor-pointer" onclick={nextScreen}>Next</Button>
	</div>
</div>
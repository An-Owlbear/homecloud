<script lang="ts">
	import { Navbar, NavBrand, NavHamburger, NavLi, NavUl } from "flowbite-svelte";
	import { ArrowLeftToBracketOutline, ColumnOutline, SearchOutline, ShoppingBagSolid } from "flowbite-svelte-icons";

	let { children } = $props();
</script>

<Navbar>
	<NavBrand href="/store" class="text-gray-700 gap-2">
		<ShoppingBagSolid />
		<span class="self-center whitespace-nowrap text-xl font-semibold">App Store</span>
	</NavBrand>
	<NavHamburger />
	<NavUl>
		<NavLi href="/store/categories" class="flex flex-row items-center gap-2">
			<ColumnOutline />
			<span>Categories</span>
		</NavLi>
		<NavLi href="/store/search" class="flex flex-row items-center gap-2">
			<SearchOutline />
			<span>Search Apps</span>
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
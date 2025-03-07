<script lang="ts">
	import '../app.css';
	import { Toast } from 'flowbite-svelte';
	import { slide } from 'svelte/transition';
	import { type ToastNotification, toasts } from '$lib/toasts.svelte';
	let { children } = $props();

	const onClose = (toast: ToastNotification) => {
		const index = toasts.indexOf(toast);
		toasts.splice(index, 1);
	}
</script>

{@render children()}

<div class="fixed bottom-5 left-4 z-50 pointer-events-none space-y-3 w-screen">
	{#each toasts as toast}
		<Toast class="pointer-events-auto shadow-xl drop-shadow-xl rounded-xl" transition={slide} on:close={() => onClose(toast)}>
			<span class="text-lg">Invite link copied!</span>
		</Toast>
	{/each}
</div>
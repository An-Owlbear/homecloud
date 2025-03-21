<script lang="ts">
	const { children } = $props();
</script>

<div class="w-full flex flex-col p-5 text-gray-500 bg-white rounded-lg border border-gray-200 cursor-pointer dark:hover:text-gray-300 dark:border-gray-700 dark:peer-checked:text-primary-500 peer-checked:border-primary-600 peer-checked:text-primary-600 hover:text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:bg-gray-800 dark:hover:bg-gray-700">
	{@render children?.()}
</div>

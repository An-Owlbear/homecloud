<script lang="ts">
	import { Heading } from "flowbite-svelte";
	import SettingLink from './SettingLink.svelte';
	import { ArchiveSolid, InfoCircleSolid, UsersSolid } from 'flowbite-svelte-icons';

</script>

<Heading tag="h1" class="mb-8">Settings</Heading>

<ul class="space-y-4">
	<SettingLink href="/settings/users" title="User settings" Icon={UsersSolid} />
	<SettingLink href="/settings/apps" title="App settings" Icon={ArchiveSolid}  />
	<SettingLink href="/settings/system" title="System settings" Icon={InfoCircleSolid} />
</ul>
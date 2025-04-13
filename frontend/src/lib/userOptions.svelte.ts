import type { UserOptions } from '$lib/models';
import { getUserOptions } from '$lib/api';

let userOptions = $state<UserOptions>();

export const getUserOptionsState = () => {
	const loadOptions = async () => {
		userOptions = await getUserOptions();
	}

	return {
		get options() {
			if (!userOptions) {
				throw new Error('User options must be loaded before use')
			}
			return userOptions;
		},
		loadOptions
	}
}
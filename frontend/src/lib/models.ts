export type HomecloudApp = {
	id: string,
	name: string,
	version: string,
	author: string,
	description: string
	image_url: string
	status: AppStatus
}

export enum AppStatus {
	Created = "created",
	Restarting = "restarting",
	Running = "running",
	Paused = "paused",
	Exited = "exited",
	Dead = "dead"
}

export type PackageListItem = {
	id: string
	name: string
	version: string
	author: string
	description: string
	categories:  string[]
	image_url: string
	installed: boolean
}

export type SearchParams = {
	q?: string
	category?: string
	developer?: string
}

export type User = {
	id: string
	metadata_public?: {
		roles?: string[]
	},
	traits: {
		email: string,
		name: string
	}
}

export type InviteCode = {
	code: string,
	expiry_date: Date
}

export type UpdateCheckResponse = {
	update_required: boolean
}

export type ExternalStorage = {
	name: string,
	label: string,
	size: number,
	available: number,
}

export type UserOptions = {
	user_id: string,
	completed_welcome: boolean
}

export type UpdateUserOptions = {
	completed_welcome?: boolean;
}

export type StoreHome = {
	popular_categories: string[],
	new_apps: PackageListItem[],
}
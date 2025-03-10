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
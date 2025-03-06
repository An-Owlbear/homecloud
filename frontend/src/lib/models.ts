export type HomecloudApp = {
	id: string,
	name: string,
	version: string,
	author: string,
	description: string
}

export type PackageListItem = {
	id: string
	name: string
	version: string
	author: string
	description: string
	categories:  string[]
	image_url: string
}

export type SearchParams = {
	q?: string
	category?: string
	developer?: string
}
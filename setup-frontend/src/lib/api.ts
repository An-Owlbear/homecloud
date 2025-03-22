export const registerDomain = async (domain: string) => {
	const response = await fetch('/api/v1/set_subdomain', {
		method: 'POST',
		body: JSON.stringify({ subdomain: domain }),
		headers: { 'Content-Type': 'application/json' },
	});
	if (!response.ok) {
		throw new Error(response.statusText);
	}
}

export const checkSubdomainTaken = async (domain: string): Promise<boolean> => {
	const response = await fetch('/api/v1/check_subdomain', {
		method: 'POST',
		body: JSON.stringify({ address: domain }),
		headers: { 'Content-Type': 'application/json' },
	});
	if (!response.ok) {
		throw new Error(response.statusText);
	}

	const body = await response.json() as { taken: boolean };
	return body.taken;
}
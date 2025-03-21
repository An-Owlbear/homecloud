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
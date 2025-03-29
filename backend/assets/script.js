/**
 * @param event InputEvent
 */
function toggleShowPassword(event) {
	const name = event.target.id.replace(/-show-password$/, '');

	/**
	 * @type {HTMLInputElement}
	 */
	const input = document.querySelector(`input[name="${name}"]`);
	if (event.target.checked) {
		input.type = 'text';
	} else {
		input.type = 'password';
	}
}
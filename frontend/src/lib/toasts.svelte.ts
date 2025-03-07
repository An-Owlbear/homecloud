export type ToastNotification = {
	content: string;
}

export const toasts = $state<ToastNotification[]>([])
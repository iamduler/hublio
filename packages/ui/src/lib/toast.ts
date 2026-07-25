import {
  toast as sonnerToast,
  Toaster as SonnerToaster,
  type ExternalToast,
  type ToastT,
} from "sonner";

export type ToastOptions = ExternalToast;
export type ToastId = string | number;

export const toast = {
  success(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.success(message, options);
  },
  error(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.error(message, options);
  },
  info(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.info(message, options);
  },
  warning(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.warning(message, options);
  },
  message(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.message(message, options);
  },
  loading(message: string, options?: ToastOptions): ToastId {
    return sonnerToast.loading(message, options);
  },
  dismiss(id?: ToastId): void {
    sonnerToast.dismiss(id);
  },
  promise: sonnerToast.promise.bind(sonnerToast) as typeof sonnerToast.promise,
} as const;

export type { ToastT };
export { SonnerToaster as AppToaster };

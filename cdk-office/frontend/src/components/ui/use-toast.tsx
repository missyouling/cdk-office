import { toast as sonnerToast } from "sonner"

type ToastProps = {
  title?: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  duration?: number
}

export const toast = {
  success: ({ title, description, duration = 4000 }: ToastProps) => {
    return sonnerToast.success(title, {
      description,
      duration,
    })
  },
  error: ({ title, description, duration = 4000 }: ToastProps) => {
    return sonnerToast.error(title, {
      description,
      duration,
    })
  },
  info: ({ title, description, duration = 4000 }: ToastProps) => {
    return sonnerToast.info(title, {
      description,
      duration,
    })
  },
  warning: ({ title, description, duration = 4000 }: ToastProps) => {
    return sonnerToast.warning(title, {
      description,
      duration,
    })
  },
  loading: ({ title, description }: ToastProps) => {
    return sonnerToast.loading(title, {
      description,
    })
  },
  dismiss: (toastId?: string | number) => {
    return sonnerToast.dismiss(toastId)
  },
}

// 简化版本，也可以直接使用
export const showToast = {
  success: (message: string) => toast.success({ title: message }),
  error: (message: string) => toast.error({ title: message }),
  info: (message: string) => toast.info({ title: message }),
  warning: (message: string) => toast.warning({ title: message }),
  loading: (message: string) => toast.loading({ title: message }),
}
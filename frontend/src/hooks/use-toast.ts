// Re-export the toast functionality from the components/ui
export { useToast, toast } from '@/components/ui/hooks/use-toast'

// Helper functions for commonly used toast types
export const success = (description: string, title: string = "成功") => {
  const { toast } = require('@/components/ui/hooks/use-toast')
  return toast({
    title,
    description,
    variant: "default",
  })
}

export const error = (description: string, title: string = "エラー") => {
  const { toast } = require('@/components/ui/hooks/use-toast')
  return toast({
    title,
    description,
    variant: "destructive",
  })
}

export const info = (description: string, title: string = "情報") => {
  const { toast } = require('@/components/ui/hooks/use-toast')
  return toast({
    title,
    description,
    variant: "default",
  })
}

export const warning = (description: string, title: string = "警告") => {
  const { toast } = require('@/components/ui/hooks/use-toast')
  return toast({
    title,
    description,
    variant: "destructive",
  })
}
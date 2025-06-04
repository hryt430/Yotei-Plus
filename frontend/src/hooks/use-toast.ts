export { useToast, toast } from '@/components/ui/hooks/use-toast'

// トーストに使うヘルパー関数
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
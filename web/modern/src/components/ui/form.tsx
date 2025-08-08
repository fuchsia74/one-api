import * as React from 'react'
import { cn } from '@/lib/utils'
import { FormProvider, useFormContext } from 'react-hook-form'

export const Form = FormProvider

export function FormItem({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('space-y-2', className)} {...props} />
}
export function FormLabel({ className, ...props }: React.LabelHTMLAttributes<HTMLLabelElement>) {
  return <label className={cn('text-sm font-medium leading-none', className)} {...props} />
}
export function FormControl({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('space-y-2', className)} {...props} />
}
export function FormMessage({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  const { formState } = useFormContext()
  return <p className={cn('text-xs text-destructive', className)} {...props}>{formState.errors?.message as any}</p>
}
export function FormField<TFieldValues extends Record<string, any>>(props: {
  name: keyof TFieldValues & string
  control: any
  render: (args: { field: any }) => React.ReactNode
}) {
  const { control, name, render } = props
  const { register } = useFormContext()
  return <>{render({ field: register(name) })}</>
}
